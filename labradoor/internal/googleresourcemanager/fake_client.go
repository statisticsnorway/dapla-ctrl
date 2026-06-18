package googleresourcemanager

import "context"

type FakeCloudResourceManager struct{}

func NewFake() *FakeCloudResourceManager {
	return &FakeCloudResourceManager{}
}

func (f *FakeCloudResourceManager) AddBindings(
	ctx context.Context,
	projectID string,
	member string,
	roles ...string,
) error {
	return nil
}

func (f *FakeCloudResourceManager) RemoveMember(
	ctx context.Context,
	projectID string,
	member string,
	roles ...string,
) error {
	return nil
}
