package gcpresources

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
)

type ResourceManager interface {
	GetOrCreateFolder(ctx context.Context, displayName, parent string) (folderID string, err error)
	GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (tagValueName string, err error)
	TagFolder(ctx context.Context, folderID, tagValueName string) error
}

type GoogleResourceManager struct {
	client *cloudresourcemanager.Service
}

func NewGoogleResourceManager(ctx context.Context) (ResourceManager, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("create cloud resource manager client: %w", err)
	}
	return &GoogleResourceManager{client: svc}, nil
}

func (g *GoogleResourceManager) GetOrCreateFolder(ctx context.Context, displayName, parent string) (string, error) {
	listResp, err := g.client.Folders.List().Parent(parent).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("list folders under %q: %w", parent, err)
	}
	for _, f := range listResp.Folders {
		if f.DisplayName == displayName {
			return folderID(f.Name), nil
		}
	}
	if _, err := g.client.Folders.Create(&cloudresourcemanager.Folder{
		DisplayName: displayName,
		Parent:      parent,
	}).Context(ctx).Do(); err != nil {
		return "", fmt.Errorf("create folder %q under %q: %w", displayName, parent, err)
	}
	var folder cloudresourcemanager.Folder
	return folderID(folder.Name), nil
}

func (g *GoogleResourceManager) GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (string, error) {
	listResp, err := g.client.TagValues.List().Parent(tagKeyNamespacedName).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("list tag values for key %q: %w", tagKeyNamespacedName, err)
	}
	for _, tv := range listResp.TagValues {
		if tv.ShortName == teamSlug {
			return tv.NamespacedName, nil
		}
	}
	if _, err := g.client.TagValues.Create(&cloudresourcemanager.TagValue{
		Parent:    tagKeyNamespacedName,
		ShortName: teamSlug,
	}).Context(ctx).Do(); err != nil {
		return "", fmt.Errorf("create tag value %q under %q: %w", teamSlug, tagKeyNamespacedName, err)
	}
	var tagValue cloudresourcemanager.TagValue
	return tagValue.NamespacedName, nil
}

func (g *GoogleResourceManager) TagFolder(ctx context.Context, fID, tagValueName string) error {
    folderResource := fmt.Sprintf("//cloudresourcemanager.googleapis.com/folders/%s", fID)
    _, err := g.client.TagBindings.Create(&cloudresourcemanager.TagBinding{
        Parent:                 folderResource,
        TagValueNamespacedName: tagValueName,
    }).Context(ctx).Do()
    if isAlreadyExists(err) {
        return nil
    }
    return err
}

func isAlreadyExists(err error) bool {
	var gErr *googleapi.Error
	return errors.As(err, &gErr) && gErr.Code == http.StatusConflict
}

func folderID(name string) string {
	return strings.TrimPrefix(name, "folders/")
}
