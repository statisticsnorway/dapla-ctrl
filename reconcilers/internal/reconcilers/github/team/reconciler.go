package team

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v88/github"
	"github.com/sirupsen/logrus"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/iterator"
	"github.com/statisticsnorway/dapla-ctrl/api/pkg/apiclient/protoapi"
)

const (
	reconcilerName = "github:team"

	configTeamAllowlistKey = "teamAllowlist"
	configTeamPrefixKey    = "teamPrefix"
)

type reconciler struct {
	teamAllowlist []string
	teamPrefix    string
	org           string
	teamsClient   *github.TeamsService
}

type optFunc func(*reconciler)

func New(ctx context.Context, org string, appId, installationId int64, privateKeyFile string, opts ...optFunc) (*reconciler, error) {
	r := &reconciler{
		org: org,
	}

	for _, opt := range opts {
		opt(r)
	}

	tr, err := ghinstallation.NewKeyFromFile(http.DefaultTransport, appId, installationId, privateKeyFile)
	if err != nil {
		return nil, err
	}

	httpClient := &http.Client{
		Timeout:   10 * time.Second,
		Transport: tr,
	}

	client, err := github.NewClient(github.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}

	r.teamsClient = client.Teams

	return r, nil
}

func (r *reconciler) Configuration() *protoapi.NewReconciler {
	return &protoapi.NewReconciler{
		Name:        r.Name(),
		DisplayName: "GitHub Team",
		Description: "Create GitHub teams and sync them with Entra ID",
		MemberAware: true,
		Config: []*protoapi.ReconcilerConfigSpec{
			{
				Key:         configTeamAllowlistKey,
				DisplayName: "Team whitelist",
				Description: "Comma-separated list of teams to sync to GitHub",
				Secret:      false,
			},
			{
				Key:         configTeamPrefixKey,
				DisplayName: "Team prefix",
				Description: "Prefix to add to GitHub teams",
				Secret:      false,
			},
		},
	}
}

func (r *reconciler) Name() string {
	return reconcilerName
}

func (r *reconciler) Reconcile(ctx context.Context, client *apiclient.APIClient, daplaTeam *protoapi.Team, log logrus.FieldLogger) error {
	if err := r.updateConfig(ctx, client); err != nil {
		return fmt.Errorf("error getting reconciler config: %w", err)
	}

	if !slices.Contains(r.teamAllowlist, daplaTeam.Slug) {
		return nil
	}

	// Iterate through all the groups in the team, and reconcile them one by one
	groupsIt := iterator.New(ctx, 100, func(limit, offset int64) (*protoapi.ListTeamGroupsResponse, error) {
		return client.Teams().Groups(ctx, &protoapi.ListTeamGroupsRequest{
			Slug: daplaTeam.Slug,
		})
	})

	for groupsIt.Next() {
		group := groupsIt.Value().Group
		if group.ExternalId == nil {
			return fmt.Errorf("group %q missing external id", group.Name)
		}
		if err := r.reconcileGroup(ctx, group.Name, *group.ExternalId); err != nil {
			return fmt.Errorf("reconcile group %q: %w", group.Name, err)
		}
	}

	return nil
}

func (r *reconciler) reconcileGroup(ctx context.Context, groupName, entraIdGroupId string) error {
	team, err := r.getOrCreateGitHubTeam(ctx, groupName)
	if err != nil {
		return err
	}

	groups, _, err := r.teamsClient.ListIDPGroupsForTeamBySlug(ctx, r.org, *team.Slug)
	if err != nil {
		return err
	}

	if len(groups.Groups) == 1 && *(groups.Groups[0].GroupID) == entraIdGroupId {
		return nil
	}

	_, _, err = r.teamsClient.CreateOrUpdateIDPGroupConnectionsBySlug(ctx, r.org, *team.Slug, github.IDPGroupList{
		Groups: []*github.IDPGroup{
			{
				GroupID:          &entraIdGroupId,
				GroupName:        &groupName,
				GroupDescription: new("Configured by Dapla API"),
			},
		},
	})

	return err
}

func (r *reconciler) getOrCreateGitHubTeam(ctx context.Context, groupName string) (*github.Team, error) {
	teamSlug := r.teamPrefix + groupName
	team, _, err := r.teamsClient.GetTeamBySlug(ctx, r.org, teamSlug)
	if err == nil {
		return team, nil
	}
	if githubError, ok := errors.AsType[*github.ErrorResponse](err); ok && githubError.Response.StatusCode == http.StatusNotFound {
		team, _, err := r.teamsClient.CreateTeam(ctx, r.org, github.NewTeam{Name: teamSlug})
		return team, err
	}
	return nil, err
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
		case configTeamAllowlistKey:
			whitelist := strings.Split(c.Value, ",")
			if !slices.Equal(r.teamAllowlist, whitelist) {
				r.teamAllowlist = whitelist
			}
		case configTeamPrefixKey:
			if r.teamPrefix != c.Value {
				r.teamPrefix = c.Value
			}
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
