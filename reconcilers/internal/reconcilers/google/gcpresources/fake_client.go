package gcpresources

import "context"

// FakeResourceManager is an in-memory ResourceManager for local development and testing.
type FakeResourceManager struct {
	TagValues map[string]string // tagKey+teamSlug -> tagValueName
	Folders   map[string]string // "parent/displayName" -> folderID
	Tags      map[string]string // folderID -> tagValueName
}

func NewFakeResourceManager() *FakeResourceManager {
	return &FakeResourceManager{
		TagValues: make(map[string]string),
		Folders:   make(map[string]string),
		Tags:      make(map[string]string),
	}
}

func (f *FakeResourceManager) GetTagValue(_ context.Context, tagKeyNamespacedName, teamSlug string) (string, error) {
	key := tagKeyNamespacedName
	if v, ok := f.TagValues[key]; ok {
		return v, nil
	}
	return "", ErrNotFound
}

func (f *FakeResourceManager) CreateTagValue(_ context.Context, tagKeyNamespacedName, teamSlug string) (string, error) {
	key := tagKeyNamespacedName
	f.TagValues[key] = teamSlug
	return key, nil
}

func (f *FakeResourceManager) GetFolder(_ context.Context, displayName, parent string) (string, error) {
	key := parent + "/" + displayName
	if id, ok := f.Folders[key]; ok {
		return id, nil
	}
	return "", ErrNotFound
}

func (f *FakeResourceManager) CreateFolder(_ context.Context, displayName, parent string) (string, error) {
	key := parent + "/" + displayName
	id := "fake-folder-" + displayName
	f.Folders[key] = id
	return id, nil
}

func (f *FakeResourceManager) TagFolder(_ context.Context, folderID, tagValueName string) error {
	f.Tags[folderID] = tagValueName
	return nil
}
