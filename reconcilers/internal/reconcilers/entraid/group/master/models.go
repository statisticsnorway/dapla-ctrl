package master

import (
	"context"

	"github.com/sirupsen/logrus"
)

type User struct {
	ExternalId string
	Email      string
}

type Group struct {
	ExternalId string
	Name       string
}

type MemberMaster interface {
	AddUsers(ctx context.Context, group Group, localOnlyUsers, remoteOnlyUsers []User, log logrus.FieldLogger) error
	RemoveUsers(ctx context.Context, group Group, localOnlyUsers, remoteOnlyUsers []User, log logrus.FieldLogger) error
	Name() string
}
