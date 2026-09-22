package atlantis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"text/template"

	googlecreds "golang.org/x/oauth2/google"

	container "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/storage"
	"github.com/google/go-cmp/cmp"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	admindirectory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/transport"

	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

const (
	reconcilerName = "atlantis"

	webhookSecretKey = "gh-webhook-secret"
	reposYamlKey     = "repos.yaml"

	wiAnnotationKey = "iam.gke.io/gcp-service-account"

	namespaceConfigKey           = "namespace"
	atlantisProjectConfigKey     = "atlantis_project"
	atlantisBaseDomainConfigKey  = "atlantis_base_domain"
	atlantisImageConfigKey       = "atlantis_image"
	memberGroupsConfigKey        = "member_groups"
	managerGroupsConfigKey       = "manager_groups"
	clusterResourceNameConfigKey = "cluster_resource_name"
	teamAllowListConfigKey       = "team_allowlist"
	tfStateProjectsConfigKey     = "tfstate_projects"
)

type groupRole string

const (
	member  groupRole = "MEMBER"
	manager groupRole = "MANAGER"
)

var (
	defaultDiskSize resource.Quantity = resource.MustParse("10Gi")
)

//go:embed repos.yaml
var defaultRepoConfig string

type reconciler struct {
	storageClient   *storage.Client
	serviceAccounts *serviceaccounts.Client
	members         *admindirectory.MembersService
	folders         *resourcemanager.FoldersClient
	clusterManager  *container.ClusterManagerClient

	knServices servingv1.ServingV1Interface
	k8sClient  kubernetes.Interface

	knativeServiceTemplate *template.Template

	config reconcilerConfig
}

type reconcilerConfig struct {
	tfstateProjects map[string]string

	clusterResourceName string

	memberGroups  []string
	managerGroups []string

	atlantisProject   string
	atlantisImage     string
	atlantisNamespace string

	teamAllowlist []string
}

type optFunc func(*reconciler)

func New(ctx context.Context, googleServices *google.Services, opts ...optFunc) (*reconciler, error) {
	r := new(reconciler)

	if googleServices != nil {
		r.storageClient = googleServices.Storage
		r.serviceAccounts = googleServices.ServiceAccounts
		r.folders = googleServices.Folders
		r.members = googleServices.AdminDirectory.Members
		r.clusterManager = googleServices.ClusterManager
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.storageClient == nil || r.serviceAccounts == nil || r.members == nil || r.folders == nil || r.clusterManager == nil {
		return nil, errors.New("one or more google clients are nil, all need to be supplied")
	}

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "Atlantis",
		Description: "Create and manage team Atlantis instances",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         namespaceConfigKey,
				DisplayName: "Namespace",
				Description: "The namespace where the atlantis resources should be deployed",
			},
			{
				Key:         atlantisProjectConfigKey,
				DisplayName: "Atlantis Project ID",
				Description: "The GCP project of the atlantis cluster",
			},
			{
				Key:         atlantisBaseDomainConfigKey,
				DisplayName: "Atlantis base domain",
				Description: "Base domain for atlantis ingresses, appended to 'https://atlantis-name.'",
			},
			{
				Key:         atlantisImageConfigKey,
				DisplayName: "Atlantis OCI image",
				Description: "Image and tag to use for atlantis instances by default",
			},
			{
				Key:         memberGroupsConfigKey,
				DisplayName: "Atlantis Member groups",
				Description: "Google groups of which every atlantis should be a member",
			},
			{
				Key:         managerGroupsConfigKey,
				DisplayName: "Atlantis Manager groups",
				Description: "Google groups of which every atlantis should be a manager",
			},
			{
				Key:         tfStateProjectsConfigKey,
				DisplayName: "Terraform State Projects",
				Description: "Map of environment names to their respective Terraform state projects",
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	// Use allowlist to perform limited testing before full rollout
	if len(r.config.teamAllowlist) != 0 && !slices.Contains(r.config.teamAllowlist, daplaTeam.Slug) {
		return nil
	}

	if err := r.updateConfig(ctx, client); err != nil {
		return err
	}

	configResponse, err := client.Atlantis().GetTeamAtlantis(ctx, &protoapi.GetTeamAtlantisRequest{TeamSlug: daplaTeam.Slug})
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	config := configResponse.Config

	// All team atlantis instances should have their resources prefixed with "atlantis-"
	atlantisName := "atlantis-" + daplaTeam.Slug
	if config.CustomName != nil {
		atlantisName = *config.CustomName
	}

	if err := r.reconcileGcpServiceAccount(ctx, client, daplaTeam.Slug, atlantisName, r.config.atlantisNamespace); err != nil {
		return err
	}

	if err := r.reconcileBuckets(ctx, daplaTeam.Slug); err != nil {
		return err
	}

	if config.WebhookSecret == nil {
		webhookSecret, err := getOrGenerateWebhookSecret(ctx, client, daplaTeam.Slug)
		if err != nil {
			return err
		}
		config.WebhookSecret = &webhookSecret
	}

	if err := r.reconcileKubernetesResources(ctx, atlantisName, r.config.atlantisNamespace, defaultRepoConfig, config, []string{"github.com/statisticsnorway/" + daplaTeam.Slug + "-iac"}); err != nil {
		return nil
	}

	return nil
}

func (r *reconciler) reconcileKubernetesResources(ctx context.Context, name, namespace, repoConfig string, config *protoapi.AtlantisConfig, repoAllowList []string) error {
	if err := r.reconcileKubernetesServiceAccount(ctx, name, namespace); err != nil {
		return err
	}

	if err := r.reconcileKubernetesWebhookSecret(ctx, name, namespace, *config.WebhookSecret); err != nil {
		return err
	}

	if err := r.reconcileKubernetesReposConfig(ctx, name, namespace, repoConfig); err != nil {
		return err
	}

	diskSize := defaultDiskSize
	if config.DiskSize != nil {
		var err error
		diskSize, err = resource.ParseQuantity(*config.DiskSize)
		if err != nil {
			return err
		}
	}
	if err := r.reconcileKubernetesVolume(ctx, name, namespace, diskSize); err != nil {
		return err
	}

	if err := r.reconcileKnativeService(ctx, name, namespace, repoAllowList, config.CustomImage, config.Resources); err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileKubernetesServiceAccount(ctx context.Context, name, namespace string) error {
	saClient := r.k8sClient.CoreV1().ServiceAccounts(namespace)
	gcpSaName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, r.config.atlantisProject)

	wantedAnnotations := map[string]string{
		wiAnnotationKey: gcpSaName,
	}

	sa, err := saClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = saClient.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:        name,
				Annotations: wantedAnnotations,
			},
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if cmp.Equal(sa.Annotations, wantedAnnotations) {
		return nil
	}

	sa.Annotations = wantedAnnotations
	_, err = saClient.Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesWebhookSecret(ctx context.Context, name, namespace, webhookSecret string) error {
	secretsClient := r.k8sClient.CoreV1().Secrets(namespace)

	wantedData := map[string][]byte{
		webhookSecretKey: []byte(webhookSecret),
	}

	secret, err := secretsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretsClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: wantedData,
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if cmp.Equal(secret.Data, wantedData) {
		return nil
	}

	secret.Data = wantedData
	_, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesReposConfig(ctx context.Context, name, namespace, repoConfig string) error {
	configMapsClient := r.k8sClient.CoreV1().ConfigMaps(namespace)

	wantedData := map[string]string{
		reposYamlKey: repoConfig,
	}

	cm, err := configMapsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = configMapsClient.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Data: wantedData,
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if cmp.Equal(cm.Data, wantedData) {
		return nil
	}

	cm.Data = wantedData
	_, err = configMapsClient.Update(ctx, cm, metav1.UpdateOptions{})

	return err
}

func (r *reconciler) reconcileKubernetesVolume(ctx context.Context, name, namespace string, diskSize resource.Quantity) error {
	pvcClient := r.k8sClient.CoreV1().PersistentVolumeClaims(namespace)
	wantedSpec := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceRequestsStorage: diskSize,
				},
			},
		},
	}

	pvc, err := pvcClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = pvcClient.Create(ctx, wantedSpec, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if equality.Semantic.DeepDerivative(wantedSpec, pvc) {
		return nil
	}

	_, err = pvcClient.Update(ctx, wantedSpec, metav1.UpdateOptions{})
	return err
}

func getOrGenerateWebhookSecret(ctx context.Context, client *apiclient.APIClient, teamName string) (string, error) {
	cfg, err := client.Atlantis().GetTeamAtlantis(ctx, &protoapi.GetTeamAtlantisRequest{TeamSlug: teamName})
	if err != nil && status.Code(err) != codes.NotFound {
		return "", err
	} else if err == nil && cfg.Config.WebhookSecret != nil {
		return *cfg.Config.WebhookSecret, nil
	}

	randBytes := make([]byte, 128)
	_, err = rand.Read(randBytes)
	if err != nil {
		return "", err
	}
	secretToken := fmt.Sprintf("%x", sha256.Sum256(randBytes))

	if _, err := client.Atlantis().SetTeamAtlantisWebhookSecret(ctx, &protoapi.SetTeamAtlantisWebhookSecretRequest{
		TeamSlug:      teamName,
		WebhookSecret: secretToken,
	}); err != nil {
		return "", err
	}

	return secretToken, nil
}

func (r *reconciler) reconcileGcpServiceAccount(ctx context.Context, client *apiclient.APIClient, teamName, name, namespace string) error {
	sa, err := r.serviceAccounts.GetOrCreate(ctx, name, "Atlantis for team "+teamName, r.config.atlantisProject)
	if err != nil {
		return err
	}

	r.serviceAccounts.EnsureRoleBindingFunc(ctx, sa.Name, "roles/iam.workloadIdentityUser", func(b *iam.Binding) bool {
		k8sSaName := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", r.config.atlantisProject, namespace, name)
		if len(b.Members) == 1 && b.Members[0] == k8sSaName {
			return false
		}
		b.Members = []string{k8sSaName}
		return true
	})

	for _, memberGroup := range r.config.memberGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, memberGroup, member); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	for _, managerGroup := range r.config.managerGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, managerGroup, manager); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	if err != nil {
		return err
	}

	saMember := "serviceAccount:" + sa.Email

	folderResp, err := client.GcpTeamResources().ListTeamFolders(ctx, &protoapi.ListGcpTeamFoldersRequest{
		TeamSlug: teamName,
	})
	if err != nil {
		return err
	}
	for _, folder := range folderResp.Folders {
		google.EnsureRolesBindingFunc(ctx, r.folders, folder.FolderId,
			[]string{"roles/resourcemanager.projectCreator", "roles/resourcemanager.projectIamAdmin"},
			func(b *iampb.Binding) (modified bool) {
				if slices.Contains(b.Members, saMember) {
					return false
				}
				b.Members = append(b.Members, saMember)
				return true
			})
	}

	return nil
}

func (r *reconciler) ensureGroupMembership(ctx context.Context, saEmail, groupId string, role groupRole) error {
	member, err := r.members.Get(groupId, saEmail).Context(ctx).Do()
	if status.Code(err) == codes.NotFound {
		_, err := r.members.Insert(groupId, &admindirectory.Member{
			Email: saEmail,
			Role:  string(role),
		}).Context(ctx).Do()
		return err
	} else if err != nil {
		return err
	}

	if member.Role == string(role) {
		return nil
	}

	_, err = r.members.Patch(groupId, saEmail, &admindirectory.Member{Etag: member.Etag, Role: string(role)}).Context(ctx).Do()
	return err
}

func (r *reconciler) reconcileBuckets(ctx context.Context, teamName string) error {
	defaultAttrs := &storage.BucketAttrs{
		UniformBucketLevelAccess: storage.UniformBucketLevelAccess{Enabled: true},
		Location:                 "EUROPE-NORTH1",
		VersioningEnabled:        true,
		PublicAccessPrevention:   storage.PublicAccessPreventionInherited,
		Lifecycle: storage.Lifecycle{
			Rules: []storage.LifecycleRule{
				{
					Action: storage.LifecycleAction{
						Type: "Delete",
					},
					Condition: storage.LifecycleCondition{
						NumNewerVersions: 3,
					},
				},
			},
		},
	}

	for env, projectId := range r.config.tfstateProjects {
		bucketName := fmt.Sprintf("ssb-%s-tfstate-%s", teamName, env)
		bucket := r.storageClient.Bucket(bucketName)
		attrs, err := bucket.Attrs(ctx)
		if status.Code(err) == codes.NotFound {
			// Create bucket
			if err := bucket.Create(ctx, projectId, defaultAttrs); err != nil {
				return fmt.Errorf("create bucket: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("get bucket attrs: %w", err)
		}
		if equality.Semantic.DeepDerivative(defaultAttrs, attrs) {
			continue
		}
		if _, err := bucket.Update(ctx, storage.BucketAttrsToUpdate{
			UniformBucketLevelAccess: &defaultAttrs.UniformBucketLevelAccess,
			VersioningEnabled:        defaultAttrs.VersioningEnabled,
			PublicAccessPrevention:   defaultAttrs.PublicAccessPrevention,
			Lifecycle:                &defaultAttrs.Lifecycle,
		}); err != nil {
			return fmt.Errorf("update bucket attrs: %w", err)
		}
	}

	return nil
}

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	rc := reconcilerConfig{}

	for _, c := range config.Nodes {
		switch c.Key {
		case namespaceConfigKey:
			rc.atlantisNamespace = c.Value
		case atlantisProjectConfigKey:
			rc.atlantisProject = c.Value
		case atlantisImageConfigKey:
			rc.atlantisImage = c.Value
		case memberGroupsConfigKey:
			if c.Value == "" {
				continue
			}
			rc.memberGroups = strings.Split(c.Value, ",")
		case managerGroupsConfigKey:
			if c.Value == "" {
				continue
			}
			rc.managerGroups = strings.Split(c.Value, ",")
		case clusterResourceNameConfigKey:
			rc.clusterResourceName = c.Value
		case teamAllowListConfigKey:
			if c.Value == "" {
				continue
			}
			rc.teamAllowlist = strings.Split(c.Value, ",")
		case tfStateProjectsConfigKey:
			if c.Value == "" {
				continue
			}
			entries := strings.Split(c.Value, ",")
			rc.tfstateProjects = make(map[string]string, len(entries))
			for _, entry := range entries {
				pair := strings.Split(entry, ":")
				if len(pair) != 2 {
					return fmt.Errorf("invalid entry: %s", entry)
				}
				rc.tfstateProjects[pair[0]] = pair[1]
			}
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if equality.Semantic.DeepEqual(rc, r.config) {
		return nil
	}

	k8sClient, knativeClient, err := r.createKubernetesClients(ctx)
	if err != nil {
		return err
	}

	r.k8sClient = k8sClient
	r.knServices = knativeClient
	r.config = rc
	return nil
}

func (r *reconciler) createKubernetesClients(ctx context.Context) (*kubernetes.Clientset, *servingv1.ServingV1Client, error) {
	// Get cluster info
	cluster, err := r.clusterManager.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: "projects/atlantis-8205/locations/europe-north1/clusters/atlantis",
	})
	if err != nil {
		return nil, nil, err
	}

	// Extract server CA cert, base64 encoded. ClusterInfo needs it decoded
	cert := cluster.MasterAuth.ClusterCaCertificate
	ca, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		return nil, nil, err
	}

	endpoint := cluster.ControlPlaneEndpointsConfig.IpEndpointsConfig.GetPublicEndpoint()

	// Get a token source for ADC
	ts, err := googlecreds.FindDefaultCredentials(ctx)
	if err != nil {
		return nil, nil, err
	}

	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	overrides.ClusterInfo.CertificateAuthorityData = ca
	// Need to specify https, defaults to http
	overrides.ClusterDefaults.Server = "https://" + endpoint

	// Create config for kubernetes clients
	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	restCfg, err := cc.ClientConfig()
	if err != nil {
		return nil, nil, err
	}

	// We wrap the underlying HTTP transport with a "middleware" which injects
	// rquests with our ADC
	restCfg.WrapTransport = transport.TokenSourceWrapTransport(ts.TokenSource)

	k8s, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, err
	}

	knsrv, err := servingv1.NewForConfig(restCfg)
	if err != nil {
		return nil, nil, err
	}

	return k8s, knsrv, nil

}

func (r *reconciler) validateConfig() error {
	fieldErrors := make(FieldsValidationError)
	setMissing := func(key string) { fieldErrors[key] = "missing value" }
	if im := r.config.atlantisImage; im == "" {
		setMissing(atlantisImageConfigKey)
	} else if !strings.Contains(im, ":") {
		fieldErrors[atlantisImageConfigKey] = "invalid image ref, must be <image>:<tag>"
	}

	if project := r.config.atlantisProject; project == "" {
		setMissing(atlantisProjectConfigKey)
	}

	if r.config.atlantisNamespace == "" {
		setMissing(namespaceConfigKey)
	}

	if r.config.clusterResourceName == "" {
		setMissing(clusterResourceNameConfigKey)
	}

	if len(fieldErrors) == 0 {
		return nil
	}

	return fieldErrors
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
