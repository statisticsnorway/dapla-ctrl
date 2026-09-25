package atlantis

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"cloud.google.com/go/iam/apiv1/iampb"
	"cloud.google.com/go/storage"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/google"
	admindirectory "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/iam/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/api/equality"
)

func (r *reconciler) reconcileGoogleResources(ctx context.Context, client *apiclient.APIClient, teamName, name, namespace string) error {
	if err := r.reconcileGcpServiceAccount(ctx, client, teamName, name, namespace); err != nil {
		return fmt.Errorf("reconcile service account: %w", err)
	}

	if err := r.reconcileBuckets(ctx, teamName); err != nil {
		return fmt.Errorf("reconcile buckets: %w", err)
	}

	return nil
}

func (r *reconciler) reconcileGcpServiceAccount(ctx context.Context, client *apiclient.APIClient, teamName, name, namespace string) error {
	sa, err := r.serviceAccounts.GetOrCreate(ctx, name, "Atlantis for team "+teamName, r.config.AtlantisProject)
	if err != nil {
		return fmt.Errorf("get or create SA: %w", err)
	}

	if err := r.serviceAccounts.EnsureRoleBindingFunc(ctx, sa.Name, "roles/iam.workloadIdentityUser", func(b *iam.Binding) bool {
		k8sSaName := fmt.Sprintf("serviceAccount:%s.svc.id.goog[%s/%s]", r.config.AtlantisProject, namespace, name)
		if len(b.Members) == 1 && b.Members[0] == k8sSaName {
			return false
		}
		b.Members = []string{k8sSaName}
		return true
	}); err != nil {
		return fmt.Errorf("ensure wi role binding: %w", err)
	}

	for _, memberGroup := range r.config.MemberGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, memberGroup, member); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	for _, managerGroup := range r.config.ManagerGroups {
		if currentErr := r.ensureGroupMembership(ctx, sa.Email, managerGroup, manager); err != nil {
			err = errors.Join(err, currentErr)
		}
	}
	if err != nil {
		return fmt.Errorf("ensure group memberships: %w", err)
	}

	saMember := "serviceAccount:" + sa.Email

	folderResp, err := client.GcpTeamResources().ListTeamFolders(ctx, &protoapi.ListGcpTeamFoldersRequest{
		TeamSlug: teamName,
	})
	if err != nil {
		return fmt.Errorf("list team folders: %w", err)
	}
	for _, folder := range folderResp.Folders {
		if err := google.EnsureRolesBindingFunc(ctx, r.folders, folder.FolderId,
			[]string{"roles/resourcemanager.projectCreator", "roles/resourcemanager.projectIamAdmin"},
			func(b *iampb.Binding) (modified bool) {
				if slices.Contains(b.Members, saMember) {
					return false
				}
				b.Members = append(b.Members, saMember)
				return true
			}); err != nil {
			return fmt.Errorf("ensure team folder iam: %w", err)
		}
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
		return fmt.Errorf("create membership: %w", err)
	}

	if member.Role == string(role) {
		return nil
	}

	_, err = r.members.Patch(groupId, saEmail, &admindirectory.Member{Etag: member.Etag, Role: string(role)}).Context(ctx).Do()
	return fmt.Errorf("update membership: %w", err)
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

	for env, projectId := range r.config.TfstateProjects {
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
