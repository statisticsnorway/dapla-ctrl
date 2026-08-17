package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
)

const (
	activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS"
)

func init() {
	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
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

	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_GRANTED", activitylog.ActivityLogEntryActionAdded, activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_REVOKED", activitylog.ActivityLogEntryActionRemoved, activityLogEntryResourceTypeArtifactRegistryGithubRepoAccess)
}

type ArtifactRegistryGithubRepoAccessGrantedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type ArtifactRegistryGithubRepoAccessRevokedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}
