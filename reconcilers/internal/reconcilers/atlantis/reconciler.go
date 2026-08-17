package atlantis

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google/serviceaccounts"
	cloudidentity "google.golang.org/api/cloudidentity/v1beta1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	reconcilerName = "atlantis"
)

type reconciler struct {
	tfstateProjects map[string]string
	storageClient   *storage.Client
	serviceAccounts *serviceaccounts.Client
	memberships     *cloudidentity.GroupsMembershipsService
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
	if err := r.reconcileServiceAccount(ctx, daplaTeam.Slug); err != nil {
		return err
	}
	if err := r.reconcileBuckets(ctx, daplaTeam.Slug); err != nil {
		return err
	}
	return nil
}
func (r *reconciler) reconcileServiceAccount(ctx context.Context, teamName string) error {
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
