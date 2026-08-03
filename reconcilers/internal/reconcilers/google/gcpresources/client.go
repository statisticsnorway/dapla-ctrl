package gcpresources

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/api/cloudresourcemanager/v3"
	"google.golang.org/api/googleapi"
)

type GcpClient interface {
	GetOrCreateFolder(ctx context.Context, displayName, parent string) (folderID string, err error)
	GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (tagValueName string, err error)
	TagFolder(ctx context.Context, folderID, tagValueName string) error
	DeleteFolder(ctx context.Context, folderID string) error
}

type googleGcpClient struct {
	client *cloudresourcemanager.Service
}

func NewGoogleGcpClient(ctx context.Context) (GcpClient, error) {
	svc, err := cloudresourcemanager.NewService(ctx)
	if err != nil {
		return nil, fmt.Errorf("create cloud resource manager client: %w", err)
	}
	return &googleGcpClient{client: svc}, nil
}

func (g *googleGcpClient) GetOrCreateFolder(ctx context.Context, displayName, parent string) (string, error) {
    findFolder := func() (string, bool, error) {
        listResp, err := g.client.Folders.List().Parent(parent).Context(ctx).Do()
        if err != nil {
            return "", false, fmt.Errorf("list folders under %q: %w", parent, err)
        }
        for _, f := range listResp.Folders {
            if f.DisplayName == displayName {
                return folderID(f.Name), true, nil
            }
        }
        return "", false, nil
    }

    if id, found, err := findFolder(); err != nil || found {
        return id, err
    }

    if _, err := g.client.Folders.Create(&cloudresourcemanager.Folder{
        DisplayName: displayName,
        Parent:      parent,
    }).Context(ctx).Do(); err != nil {
        return "", fmt.Errorf("create folder %q under %q: %w", displayName, parent, err)
    }

    id, found, err := findFolder()
    if err != nil {
        return "", err
    }
    if !found {
        return "", fmt.Errorf("folder %q not found after creation — will retry next cycle", displayName)
    }
    return id, nil
}

func (g *googleGcpClient) GetOrCreateTagValue(ctx context.Context, tagKeyNamespacedName, teamSlug string) (string, error) {
    findTagValue := func() (string, bool, error) {
        listResp, err := g.client.TagValues.List().Parent(tagKeyNamespacedName).Context(ctx).Do()
        if err != nil {
            return "", false, fmt.Errorf("list tag values for key %q: %w", tagKeyNamespacedName, err)
        }
        for _, tv := range listResp.TagValues {
            if tv.ShortName == teamSlug {
                return tv.NamespacedName, true, nil
            }
        }
        return "", false, nil
    }

    if name, found, err := findTagValue(); err != nil || found {
        return name, err
    }

    if _, err := g.client.TagValues.Create(&cloudresourcemanager.TagValue{
        Parent:    tagKeyNamespacedName,
        ShortName: teamSlug,
    }).Context(ctx).Do(); err != nil {
        return "", fmt.Errorf("create tag value %q under %q: %w", teamSlug, tagKeyNamespacedName, err)
    }

    name, found, err := findTagValue()
    if err != nil {
        return "", err
    }
    if !found {
        return "", fmt.Errorf("tag value %q not found after creation — will retry next cycle", teamSlug)
    }
    return name, nil
}

func (g *googleGcpClient) TagFolder(ctx context.Context, fID, tagValueName string) error {
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

func (g *googleGcpClient) DeleteFolder(ctx context.Context, fID string) error {
    _, err := g.client.Folders.Delete(fmt.Sprintf("folders/%s", fID)).Context(ctx).Do()
    if isNotFound(err) {
        return nil
    }
    return err
}

func isNotFound(err error) bool {
	var gErr *googleapi.Error
	return errors.As(err, &gErr) && gErr.Code == http.StatusNotFound
}

func isAlreadyExists(err error) bool {
	var gErr *googleapi.Error
	return errors.As(err, &gErr) && gErr.Code == http.StatusConflict
}

func folderID(name string) string {
	const prefix = "folders/"
	if len(name) > len(prefix) {
		return name[len(prefix):]
	}
	return name
}
