package gcpresources

import (
	"context"
	"fmt"
	"strings"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ResourceManager interface {
	GetOrCreateFolder(ctx context.Context, displayName, parent string) (folderID string, err error)
	GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (tagValueName string, err error)
	TagFolder(ctx context.Context, folderID, tagValueName string) error
}

type GoogleResourceManager struct {
	folders     *resourcemanager.FoldersClient
	tagValues   *resourcemanager.TagValuesClient
	tagBindings *resourcemanager.TagBindingsClient
}

func NewGoogleResourceManager(ctx context.Context) (ResourceManager, error) {
	folders, err := resourcemanager.NewFoldersClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create folders client: %w", err)
	}

	tagValues, err := resourcemanager.NewTagValuesClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tag values client: %w", err)
	}

	tagBindings, err := resourcemanager.NewTagBindingsClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("create tag bindings client: %w", err)
	}

	return &GoogleResourceManager{
		folders:     folders,
		tagValues:   tagValues,
		tagBindings: tagBindings,
	}, nil
}

func (g *GoogleResourceManager) GetOrCreateFolder(ctx context.Context, displayName, parent string) (string, error) {
	it := g.folders.ListFolders(ctx, &resourcemanagerpb.ListFoldersRequest{Parent: parent})
	for folder, err := range it.All() {
		if err != nil {
			return "", fmt.Errorf("list folders under %q: %w", parent, err)
		}
		if folder.DisplayName == displayName {
			return folderID(folder.Name), nil
		}
	}

	_, err := g.folders.CreateFolder(ctx, &resourcemanagerpb.CreateFolderRequest{
		Folder: &resourcemanagerpb.Folder{
			DisplayName: displayName,
			Parent:      parent,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create folder %q under %q: %w", displayName, parent, err)
	}
	it = g.folders.ListFolders(ctx, &resourcemanagerpb.ListFoldersRequest{Parent: parent})
	for folder, err := range it.All() {
		if err != nil {
			return "", fmt.Errorf("list folders under %q: %w", parent, err)
		}
		if folder.DisplayName == displayName {
			return folderID(folder.Name), nil
		}
	}
	return "", fmt.Errorf("folder %q not found after creation — will retry next cycle", displayName)
}

func (g *GoogleResourceManager) GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (string, error) {
	it := g.tagValues.ListTagValues(ctx, &resourcemanagerpb.ListTagValuesRequest{Parent: tagKeyNamespacedName})
	for tagValue, err := range it.All() {
		if err != nil {
			return "", fmt.Errorf("list tag values for key %q: %w", tagKeyNamespacedName, err)
		}
		if tagValue.ShortName == teamSlug {
			return tagValue.NamespacedName, nil
		}
	}

	_, err := g.tagValues.CreateTagValue(ctx, &resourcemanagerpb.CreateTagValueRequest{
		TagValue: &resourcemanagerpb.TagValue{
			Parent:    tagKeyNamespacedName,
			ShortName: teamSlug,
		},
	})
	if err != nil {
		return "", fmt.Errorf("create tag value %q under %q: %w", teamSlug, tagKeyNamespacedName, err)
	}
	it = g.tagValues.ListTagValues(ctx, &resourcemanagerpb.ListTagValuesRequest{Parent: tagKeyNamespacedName})
	for tagValue, err := range it.All() {
		if err != nil {
			return "", fmt.Errorf("list tag values for key %q: %w", tagKeyNamespacedName, err)
		}
		if tagValue.ShortName == teamSlug {
			return tagValue.NamespacedName, nil
		}
	}
	return "", fmt.Errorf("tag value %q not found after creation — will retry next cycle", teamSlug)
}

func (g *GoogleResourceManager) TagFolder(ctx context.Context, fID, tagValueName string) error {
	folderResource := fmt.Sprintf("//cloudresourcemanager.googleapis.com/folders/%s", fID)
	_, err := g.tagBindings.CreateTagBinding(ctx, &resourcemanagerpb.CreateTagBindingRequest{
		TagBinding: &resourcemanagerpb.TagBinding{
			Parent:                 folderResource,
			TagValueNamespacedName: tagValueName,
		},
	})
	if status.Code(err) == codes.AlreadyExists {
		return nil
	}
	return err
}

func folderID(name string) string {
	return strings.TrimPrefix(name, "folders/")
}
