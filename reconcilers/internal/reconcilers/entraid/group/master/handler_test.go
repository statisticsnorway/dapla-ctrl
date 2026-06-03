package master_test

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/reconcilers/internal/reconcilers/entraid/group/master"
)

type FakeMaster struct {
	name string
}

func (m FakeMaster) AddUsers(ctx context.Context, group master.Group, localOnlyUsers []master.User, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	panic("unimplemented")
}

func (m FakeMaster) RemoveUsers(ctx context.Context, group master.Group, localOnlyUsers []master.User, remoteOnlyUsers []master.User, log logrus.FieldLogger) error {
	panic("unimplemented")
}

func (m FakeMaster) Name() string {
	return m.name
}

func TestHandlerWithOverrides(t *testing.T) {
	fakeDefault := FakeMaster{"default"}
	fakeEntraid := FakeMaster{"entraid"}
	fakeDatabase := FakeMaster{"database"}

	handler, err := master.NewHandler("team:my-team:database,group:my-team-devs:entraid", fakeDefault, fakeEntraid, fakeDatabase)
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		Description string
		Team        string
		Group       string
		Expected    string
	}{
		{
			Description: "my-team-admins should have database as master",
			Team:        "my-team",
			Group:       "my-team-admins",
			Expected:    "database",
		},
		{
			Description: "my-team-devs should have entraid as master",
			Team:        "my-team",
			Group:       "my-team-devs",
			Expected:    "entraid",
		},
		{
			Description: "Other teams/groups should have default as master",
			Team:        "other-team",
			Group:       "other-team-devs",
			Expected:    "default",
		},
	}

	for _, tc := range cases {
		t.Run(tc.Description, func(t *testing.T) {
			if m := handler.GetMasterFor(tc.Team, tc.Group); m.Name() != tc.Expected {
				t.Errorf("expected %q, got %q", tc.Expected, m.Name())
			}
		})
	}
}
