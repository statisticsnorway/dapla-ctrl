package atlantis

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"text/template"

	container "cloud.google.com/go/container/apiv1"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	admindirectory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/client-go/kubernetes"

	servingv1 "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

const (
	reconcilerName = "atlantis"

	webhookSecretKey = "gh-webhook-secret" // #nosec G101 -- This is a config key, not a credential
	reposYamlKey     = "repos.yaml"

	wiAnnotationKey = "iam.gke.io/gcp-service-account"

	namespaceConfigKey           = "namespace"
	atlantisProjectConfigKey     = "atlantis_project"
	atlantisBaseDomainConfigKey  = "atlantis_base_domain"
	atlantisImageConfigKey       = "atlantis_image"
	memberGroupsConfigKey        = "member_groups"
	managerGroupsConfigKey       = "manager_groups"
	clusterResourceNameConfigKey = "cluster_resource_name"
	tfStateProjectsConfigKey     = "tfstate_projects"
	githubAppIdConfigKey         = "github_app_id"
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

//go:embed defaultservice.yaml.gotmpl
var defaultKnativeServiceTemplate string

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

	atlantisProject    string
	atlantisImage      string
	atlantisNamespace  string
	atlantisBaseDomain string

	githubAppId string
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

	if r.knativeServiceTemplate == nil {
		tmpl, err := template.New("").Parse(defaultKnativeServiceTemplate)
		if err != nil {
			return nil, err
		}
		r.knativeServiceTemplate = tmpl
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
			{
				Key:         githubAppIdConfigKey,
				DisplayName: "GitHub App Id",
				Description: "The GitHub App Id the Atlantis should use",
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if err := r.updateConfig(ctx, client); err != nil {
		return err
	}

	configResponse, err := client.Atlantis().GetTeamAtlantis(ctx, &protoapi.GetTeamAtlantisRequest{TeamSlug: daplaTeam.Slug})
	if err != nil && status.Code(err) == codes.NotFound {
		log.Debug("skipping team as they have no atlantis config")
		return nil
	} else if err != nil {
		return err
	}

	config := configResponse.Config

	// All team atlantis instances should have their resources prefixed with "atlantis-"
	// unless they have a custom name
	atlantisName := "atlantis-" + daplaTeam.Slug
	if config.CustomName != nil {
		atlantisName = *config.CustomName
	}

	if err := r.reconcileGoogleResources(ctx, client, daplaTeam.Slug, atlantisName, r.config.atlantisNamespace); err != nil {
		return err
	}

	if config.WebhookSecret == nil {
		config.WebhookSecret, err = createWebhookSecret(ctx, client, daplaTeam.Slug)
		if err != nil {
			return err
		}
	}

	if err := r.reconcileKubernetesResources(ctx, atlantisName, r.config.atlantisNamespace, config, []string{"github.com/statisticsnorway/" + daplaTeam.Slug + "-iac"}); err != nil {
		return err
	}

	return nil
}

func createWebhookSecret(ctx context.Context, client *apiclient.APIClient, teamName string) (*string, error) {
	randBytes := make([]byte, 128)
	_, err := rand.Read(randBytes)
	if err != nil {
		return nil, err
	}
	secretToken := fmt.Sprintf("%x", sha256.Sum256(randBytes))

	if _, err := client.Atlantis().SetTeamAtlantisWebhookSecret(ctx, &protoapi.SetTeamAtlantisWebhookSecretRequest{
		TeamSlug:      teamName,
		WebhookSecret: secretToken,
	}); err != nil {
		return nil, err
	}

	return &secretToken, nil
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
		case atlantisBaseDomainConfigKey:
			rc.atlantisBaseDomain = c.Value
		case githubAppIdConfigKey:
			rc.githubAppId = c.Value
		default:
			return fmt.Errorf("unknown config key %q", c.Key)
		}
	}

	if err := rc.Validate(); err != nil {
		return err
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

func (c reconcilerConfig) Validate() error {
	fieldErrors := make(FieldsValidationError)
	setMissing := func(key string) { fieldErrors[key] = "missing value" }
	if im := c.atlantisImage; im == "" {
		setMissing(atlantisImageConfigKey)
	} else if !strings.Contains(im, ":") {
		fieldErrors[atlantisImageConfigKey] = "invalid image ref, must be <image>:<tag>"
	}

	if project := c.atlantisProject; project == "" {
		setMissing(atlantisProjectConfigKey)
	}

	if c.atlantisNamespace == "" {
		setMissing(namespaceConfigKey)
	}

	if c.clusterResourceName == "" {
		setMissing(clusterResourceNameConfigKey)
	}

	if c.atlantisBaseDomain == "" {
		setMissing(atlantisBaseDomainConfigKey)
	}

	if c.githubAppId == "" {
		setMissing(githubAppIdConfigKey)
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
