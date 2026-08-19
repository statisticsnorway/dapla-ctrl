package atlantis

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"slices"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	cloudidentity "google.golang.org/api/cloudidentity/v1beta1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	knv1 "knative.dev/serving/pkg/apis/serving/v1"
	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

const (
	reconcilerName = "atlantis"

	webhookSecretKey = "gh-webhook-secret"
	reposYamlKey     = "repos.yaml"

	wiAnnotationKey = "iam.gke.io/gcp-service-account"
)

//go:embed repos.yaml
var defaultRepoConfig string

type reconciler struct {
	tfstateProjects map[string]string

	storageClient   *storage.Client
	serviceAccounts *serviceaccounts.Client
	memberships     *cloudidentity.GroupsMembershipsService

	knServices servingv1.ServiceInterface
	k8sClient  kubernetes.Interface

	atlantisProject string
	memberGroups    []string
	managerGroups   []string
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
		Config:      []*protoapi.ReconcilerConfigSpec{},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if err := r.reconcileGcpServiceAccount(ctx, daplaTeam.Slug); err != nil {
		return err
	}

	if err := r.reconcileBuckets(ctx, daplaTeam.Slug); err != nil {
		return err
	}

	webhookSecret, err := getOrGenerateWebhookSecret(ctx, client, daplaTeam.Slug)
	if err != nil {
		return err
	}

	if err := r.reconcileKubernetesResources(ctx, daplaTeam.Slug, webhookSecret, defaultRepoConfig, resource.MustParse("10Gi")); err != nil {
		return nil
	}

	return nil
}

func (r *reconciler) reconcileKubernetesResources(ctx context.Context, teamName string, webhookSecret, repoConfig string, diskSize resource.Quantity) error {
	atlantisName := "atlantis-" + teamName

	if err := r.reconcileKubernetesServiceAccount(ctx, atlantisName); err != nil {
		return err
	}
	if err := r.reconcileKubernetesSecret(ctx, atlantisName, webhookSecret); err != nil {
		return err
	}
	if err := r.reconcileKubernetesConfigMap(ctx, atlantisName, repoConfig); err != nil {
		return err
	}
	if err := r.reconcileKubernetesVolume(ctx, atlantisName, diskSize); err != nil {
		return err
	}
	return nil
}

func (r *reconciler) reconcileKubernetesSecret(ctx context.Context, name string, webhookSecret string) error {
	secretsClient := r.k8sClient.CoreV1().Secrets("default")
	secretBytes := []byte(webhookSecret)

	secret, err := secretsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = secretsClient.Create(ctx, &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Data: map[string][]byte{
				webhookSecretKey: secretBytes,
			},
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	kubernetesSecretValue := secret.Data[webhookSecretKey]
	if len(secret.Data) == 1 && bytes.Equal(kubernetesSecretValue, secretBytes) {
		return nil
	}

	secret.Data = map[string][]byte{
		webhookSecretKey: secretBytes,
	}
	_, err = secretsClient.Update(ctx, secret, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesConfigMap(ctx context.Context, name string, repoConfig string) error {
	configMapsClient := r.k8sClient.CoreV1().ConfigMaps("default")

	cm, err := configMapsClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = configMapsClient.Create(ctx, &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
			},
			Data: map[string]string{
				reposYamlKey: repoConfig,
			},
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	configMapValue := cm.Data[reposYamlKey]
	if len(cm.Data) == 1 && configMapValue == repoConfig {
		return nil
	}

	cm.Data = map[string]string{
		reposYamlKey: repoConfig,
	}
	_, err = configMapsClient.Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesServiceAccount(ctx context.Context, name string) error {
	saClient := r.k8sClient.CoreV1().ServiceAccounts("default")
	gcpSaName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", name, r.atlantisProject)

	sa, err := saClient.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = saClient.Create(ctx, &corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "default",
				Annotations: map[string]string{
					wiAnnotationKey: gcpSaName,
				},
			},
		}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	if sa.Annotations[wiAnnotationKey] == gcpSaName {
		return nil
	}

	sa.Annotations[wiAnnotationKey] = gcpSaName
	_, err = saClient.Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (r *reconciler) reconcileKubernetesVolume(ctx context.Context, name string, diskSize resource.Quantity) error {
	pvcClient := r.k8sClient.CoreV1().PersistentVolumeClaims("default")
	wantedSpec := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
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

func (r *reconciler) reconcileKnativeService(ctx context.Context, name string) error {
	env := map[string]string{
		"ATLANTIS_REPO_ALLOWLIST":                        "local.repo_allowlist",
		"ATLANTIS_GH_APP_ID":                             "",
		"ATLANTIS_GH_APP_KEY_FILE":                       "/secret/atlantis-app-key.pem",
		"ATLANTIS_WRITE_GIT_CREDS":                       "true",
		"ATLANTIS_DATA_DIR":                              "/atlantis",
		"ATLANTIS_ATLANTIS_URL":                          "",
		"ATLANTIS_PORT":                                  "4141",
		"TF_PLUGIN_CACHE_MAY_BREAK_DEPENDENCY_LOCK_FILE": "true",
		"ATLANTIS_GH_ALLOW_MERGEABLE_BYPASS_APPLY":       "true",
		"ATLANTIS_ENABLE_REGEXP_CMD":                     "true",
		"ATLANTIS_REPO_CONFIG":                           "/config/repos.yaml",
	}
	envVars := make([]corev1.EnvVar, 0, len(env)+1)
	for key, val := range env {
		envVars = append(envVars, corev1.EnvVar{
			Name:  key,
			Value: val,
		})
	}
	envVars = append(envVars, corev1.EnvVar{
		Name: "ATLANTIS_GH_WEBHOOK_SECRET",
		ValueFrom: &corev1.EnvVarSource{
			SecretKeyRef: &corev1.SecretKeySelector{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: name,
				},
				Key: webhookSecretKey,
			},
		},
	})

	probe := corev1.Probe{
		PeriodSeconds: 60,
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   "/healthz",
				Port:   4141,
				Scheme: corev1.URISchemeHTTP,
			},
		},
	}

	knService := &knv1.Service{
		Spec: knv1.ServiceSpec{
			ConfigurationSpec: knv1.ConfigurationSpec{
				Template: knv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"autoscaling.knative.dev/max-scale":                          "1",
							"autoscaling.knative.dev/scale-to-zero-pod-retention-period": "1h",
						},
					},
					Spec: knv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							ServiceAccountName: name,
							SecurityContext: &corev1.PodSecurityContext{
								FSGroup: new(int64(1000)),
							},
							Containers: []corev1.Container{
								{
									Image: "sdlkfskldfjsdlkfjds TODO",
									Env:   envVars,
									VolumeMounts: []corev1.VolumeMount{
										{
											Name:      "atlantis-data",
											MountPath: "/atlantis",
										},
										{
											Name:      "secret-volume",
											ReadOnly:  true,
											MountPath: "/secret",
										},
										{
											Name:      "config-volume",
											ReadOnly:  true,
											MountPath: "/config",
										},
									},
									Ports: []corev1.ContainerPort{
										{
											ContainerPort: 4141,
										},
									},
									LivenessProbe:  &probe,
									ReadinessProbe: &probe,
									Resources: corev1.ResourceRequirements{
										// TODO: investigage resourceXXX vs resource(requests/liits)XXX
										Requests: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("100m"),
											corev1.ResourceMemory: resource.MustParse("256Mi"),
										},
										Limits: corev1.ResourceList{
											corev1.ResourceCPU:    resource.MustParse("500m"),
											corev1.ResourceMemory: resource.MustParse("512Mi"), // TODO: get limit from API
										},
									},
								},
							},
							Volumes: []corev1.Volume{
								{
									Name: "atlantis-data",
									VolumeSource: corev1.VolumeSource{
										PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
											ClaimName: name,
											ReadOnly:  false,
										},
									},
								},
								{
									Name: "secret-volume",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: "dapla-team",
											Items: []corev1.KeyToPath{
												{
													Key:  "gh-key-file",
													Path: "atlantis-app-key.pem",
												},
											},
										},
									},
								},
								{
									Name: "config-volume",
									VolumeSource: corev1.VolumeSource{
										ConfigMap: &corev1.ConfigMapVolumeSource{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: name,
											},
											Items: []corev1.KeyToPath{
												{
													Key:  "repos.yaml",
													Path: "repos.yaml",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	kns, err := r.knServices.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = r.knServices.Create(ctx, &knv1.Service{}, metav1.CreateOptions{})
		return err
	} else if err != nil {
		return err
	}

	sa.Annotations[wiAnnotationKey] = gcpSaName
	_, err = saClient.Update(ctx, sa, metav1.UpdateOptions{})
	return err
	return nil
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

func (r *reconciler) reconcileGcpServiceAccount(ctx context.Context, teamName string) error {
	sa, err := r.serviceAccounts.GetOrCreate(ctx, "atlantis-"+teamName, "Atlantis for team "+teamName, r.atlantisProject)
	if err != nil {
		return err
	}

	r.serviceAccounts.EnsureRoleBindingFunc(ctx, sa.Name, "roles/iam.workloadIdentityUser", func(b *iam.Binding) bool {
		k8sSaName := fmt.Sprintf("serviceAccount:%s.svc.id.goog[default/atlantis-%s]", r.atlantisProject, teamName)
		if len(b.Members) == 1 && b.Members[0] == k8sSaName {
			return false
		}
		b.Members = []string{k8sSaName}
		return true
	})

	for _, memberGroup := range r.memberGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, memberGroup, false); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	for _, managerGroup := range r.managerGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, managerGroup, true); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	if err != nil {
		return err
	}

	return nil
}

func (r *reconciler) ensureGroupMembership(ctx context.Context, saEmail string, groupId string, manager bool) error {
	// Check if membership exists
	_, err := r.memberships.Lookup(groupId).MemberKeyId(saEmail).Context(ctx).Do()
	// Does exist (2xx response)
	if err == nil {
		// TODO: Check if roles are correct
		return nil
	}

	// Unknown error (not 2xx and not 404)
	if status.Code(err) != codes.NotFound {
		return err
	}

	roles := []*cloudidentity.MembershipRole{
		{
			Name: "MEMBER",
		},
	}
	if manager {
		roles = append(roles, &cloudidentity.MembershipRole{
			Name: "MANAGER",
		})
	}

	if _, err := r.memberships.Create(groupId, &cloudidentity.Membership{
		PreferredMemberKey: &cloudidentity.EntityKey{
			Id: saEmail,
		},
		Roles: roles,
	}).Context(ctx).Do(); err != nil {
		return err
	}

	return nil
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

func (r *reconciler) updateConfig(ctx context.Context, client *apiclient.APIClient) error {
	config, err := client.Reconcilers().Config(ctx, &protoapi.ConfigReconcilerRequest{
		ReconcilerName: r.Name(),
	})
	if err != nil {
		return fmt.Errorf("get reconciler config: %w", err)
	}

	for _, c := range config.Nodes {
		switch c.Key {
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	return nil
}

func (r *reconciler) Delete(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	log.Debug("Executing some action to delete the resource owned by this reconciler")

	return nil
}
