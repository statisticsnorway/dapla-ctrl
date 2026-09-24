package google

import (
	"context"

	artifactregistry "cloud.google.com/go/artifactregistry/apiv1"
	budgets "cloud.google.com/go/billing/budgets/apiv1"
	container "cloud.google.com/go/container/apiv1"
	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	serviceusage "cloud.google.com/go/serviceusage/apiv1"
	"cloud.google.com/go/storage"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	admindirectory "google.golang.org/api/admin/directory/v1"
)

type Services struct {
	Projects            *resourcemanager.ProjectsClient
	Folders             *resourcemanager.FoldersClient
	TagValues           *resourcemanager.TagValuesClient
	TagBindings         *resourcemanager.TagBindingsClient
	ServiceUsage        *serviceusage.Client
	CloudBudget         *budgets.BudgetClient
	NotificationChannel *monitoring.NotificationChannelClient
	ArtifactRegistry    *artifactregistry.Client
	ServiceAccounts     *serviceaccounts.Client
	Storage             *storage.Client
	AdminDirectory      *admindirectory.Service
	ClusterManager      *container.ClusterManagerClient
}

func New(ctx context.Context) (*Services, error) {
	s := new(Services)
	var err error

	s.Projects, err = resourcemanager.NewProjectsClient(ctx)
	if err != nil {
		return nil, err
	}

	s.Folders, err = resourcemanager.NewFoldersClient(ctx)
	if err != nil {
		return nil, err
	}

	s.TagValues, err = resourcemanager.NewTagValuesClient(ctx)
	if err != nil {
		return nil, err
	}

	s.TagBindings, err = resourcemanager.NewTagBindingsClient(ctx)
	if err != nil {
		return nil, err
	}

	s.ServiceUsage, err = serviceusage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	s.CloudBudget, err = budgets.NewBudgetClient(ctx)
	if err != nil {
		return nil, err
	}

	s.NotificationChannel, err = monitoring.NewNotificationChannelClient(ctx)
	if err != nil {
		return nil, err
	}

	s.ArtifactRegistry, err = artifactregistry.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	s.Storage, err = storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	s.AdminDirectory, err = admindirectory.NewService(ctx)
	if err != nil {
		return nil, err
	}

	s.ClusterManager, err = container.NewClusterManagerClient(ctx)
	if err != nil {
		return nil, err
	}

	s.ServiceAccounts, err = serviceaccounts.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	return s, nil
}
