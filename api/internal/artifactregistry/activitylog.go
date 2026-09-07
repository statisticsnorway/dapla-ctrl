package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
)

const (
	activityLogEntryResourceTypeArtifactRegistryAllowedGithubRepos activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS"
	activityLogEntryResourceTypeArtifactRegistryRepository         activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_REPOSITORY"
)

func init() {
	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryAllowedGithubRepos, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionAdded:
			return ArtifactRegistryGithubRepoAccessGrantedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Granted github repository access to artifact registry for the team"),
			}, nil
		case activitylog.ActivityLogEntryActionRemoved:
			return ArtifactRegistryGithubRepoAccessRevokedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Revoked github repository access from artifact registry for the team"),
			}, nil
		default:
			return nil, fmt.Errorf("unsupported repository activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryRepository, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionCreated:
			return ArtifactRegistryRepositoryCreatedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Create Artifact Registry repository"),
			}, nil
		default:
			return nil, fmt.Errorf("unsupported repository activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_GRANTED", activitylog.ActivityLogEntryActionAdded, activityLogEntryResourceTypeArtifactRegistryAllowedGithubRepos)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_REVOKED", activitylog.ActivityLogEntryActionRemoved, activityLogEntryResourceTypeArtifactRegistryAllowedGithubRepos)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_REPOSITORY_CREATED", activitylog.ActivityLogEntryActionCreated, activityLogEntryResourceTypeArtifactRegistryRepository)
}

type ArtifactRegistryGithubRepoAccessGrantedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type ArtifactRegistryGithubRepoAccessRevokedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type ArtifactRegistryRepositoryCreatedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}
