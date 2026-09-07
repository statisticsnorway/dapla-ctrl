package atlantis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"slices"

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
	cloudidentity "google.golang.org/api/cloudidentity/v1beta1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

const (
	reconcilerName = "atlantis"

	webhookSecretKey = "gh-webhook-secret"
	reposYamlKey     = "repos.yaml"

	wiAnnotationKey = "iam.gke.io/gcp-service-account"

	namespaceConfigKey = "namespace"
)

type groupRole string

const (
	member  groupRole = "MEMBER"
	manager groupRole = "MANAGER"
)

//go:embed repos.yaml
var defaultRepoConfig string

type reconciler struct {
	tfstateProjects map[string]string

	storageClient   *storage.Client
	serviceAccounts *serviceaccounts.Client
	memberships     *cloudidentity.GroupsMembershipsService
	members         *admindirectory.MembersService
	folders         *resourcemanager.FoldersClient

	knServices servingv1.ServiceInterface
	k8sClient  kubernetes.Interface

	atlantisProject string
	memberGroups    []string
	managerGroups   []string
}

type reconcilerConfig struct {
	Namespace string
}

type optFunc func(*reconciler)

func New(ctx context.Context, opts ...optFunc) (*reconciler, error) {
	r := &reconciler{
		tfstateProjects: make(map[string]string),
	}

	for _, opt := range opts {
		opt(r)
	}

	if r.storageClient == nil {
		storageClient, err := storage.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		r.storageClient = storageClient
	}

	if r.serviceAccounts == nil {
		serviceAccounts, err := serviceaccounts.NewClient(ctx)
		if err != nil {
			return nil, err
		}
		r.serviceAccounts = serviceAccounts
	}

	if r.memberships == nil {
		ci, err := cloudidentity.NewService(ctx)
		if err != nil {
			return nil, err
		}
		r.memberships = ci.Groups.Memberships
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
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	namespace := "default"
	atlantisName := "atlantis-" + daplaTeam.Slug

	sa, err := r.reconcileGcpServiceAccount(ctx, daplaTeam.Slug, atlantisName, namespace)
	if err != nil {
		return err
	}
	saMember := "serviceAccount:" + sa.Email

	folderResp, err := client.GcpTeamResources().ListTeamFolders(ctx, &protoapi.ListGcpTeamFoldersRequest{
		TeamSlug: daplaTeam.Slug,
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

	if err := r.reconcileBuckets(ctx, daplaTeam.Slug); err != nil {
		return err
	}

	webhookSecret, err := getOrGenerateWebhookSecret(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}

	if err := r.reconcileKubernetesResources(ctx, atlantisName, namespace, webhookSecret, defaultRepoConfig, resource.MustParse("10Gi")); err != nil {
		return nil
	}

	return nil
}

func (r *reconciler) reconcileKubernetesResources(ctx context.Context, name, namespace, webhookSecret, repoConfig string, diskSize resource.Quantity) error {

	if err := r.reconcileKubernetesServiceAccount(ctx, name, namespace); err != nil {
		return err
	}
	if err := r.reconcileKubernetesSecret(ctx, name, namespace, webhookSecret); err != nil {
		return err
	}
	if err := r.reconcileKubernetesConfigMap(ctx, name, namespace, repoConfig); err != nil {
		return err
	}
	if err := r.reconcileKubernetesVolume(ctx, name, namespace, diskSize); err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileKubernetesSecret(ctx context.Context, name, namespace, webhookSecret string) error {
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

func (r *reconciler) reconcileKubernetesConfigMap(ctx context.Context, name, namespace, repoConfig string) error {
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

	_, err = configMapsClient.Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesServiceAccount(ctx context.Context, name, namespace string) error {
	saClient := r.k8sClient.CoreV1().ServiceAccounts(namespace)
	gcpSaName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, r.atlantisProject)

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

	if slices.Equal(pvc.Spec.AccessModes, wantedSpec.Spec.AccessModes) &&
		maps.Equal(pvc.Spec.Resources.Requests, wantedSpec.Spec.Resources.Requests) {
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

func (r *reconciler) reconcileGcpServiceAccount(ctx context.Context, teamName, name, namespace string) (*iam.ServiceAccount, error) {
	sa, err := r.serviceAccounts.GetOrCreate(ctx, name, "Atlantis for team "+teamName, r.atlantisProject)
	if err != nil {
		return nil, err
	}

	r.serviceAccounts.EnsureRoleBindingFunc(ctx, sa.Name, "roles/iam.workloadIdentityUser", func(b *iam.Binding) bool {
		k8sSaName := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", r.atlantisProject, namespace, name)
		if len(b.Members) == 1 && b.Members[0] == k8sSaName {
			return false
		}
		b.Members = []string{k8sSaName}
		return true
	})

	for _, memberGroup := range r.memberGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, memberGroup, member); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	for _, managerGroup := range r.managerGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, managerGroup, manager); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	if err != nil {
		return nil, err
	}

	return sa, nil
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

	for env, projectId := range r.tfstateProjects {
		bucketName := fmt.Sprintf("ssb-%s-tfstate-%s", teamName, env)
		bucket := r.storageClient.Bucket(bucketName)
		_, err := bucket.Attrs(ctx)
		if status.Code(err) == codes.NotFound {
			// Create bucket
			if err := bucket.Create(ctx, projectId, defaultAttrs); err != nil {
				return fmt.Errorf("create bucket: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("get bucket attrs: %w", err)
		}
		// TODO: check that bucket attrs are correct
	}

	return nil
}

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) (*reconcilerConfig, error) {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return nil, fmt.Errorf("get reconciler config: %w", err)
	}

	rc := reconcilerConfig{}

	for _, c := range config.Nodes {
		switch c.Key {
		case namespaceConfigKey:
			rc.Namespace = c.Value
		default:
			return nil, fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	return &rc, nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
