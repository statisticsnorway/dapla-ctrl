package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
)

const (
	activityLogEntryResourceTypeArtifactRegistryGithubRepository activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS"
)

func init() {
	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryGithubRepository, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionAdded:
			return ArtifactRegistryGithubRepositoryAccessGrantedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Granted github repository access to artifact registry for the team"),
			}, nil
		case activitylog.ActivityLogEntryActionRemoved:
			return ArtifactRegistryGithubRepositoryAccessRevokedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Revoked github repository access to artifact registry from team"),
			}, nil

		default:
			return nil, fmt.Errorf("unsupported repository activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_GRANTED", activitylog.ActivityLogEntryActionAdded, activityLogEntryResourceTypeArtifactRegistryGithubRepository)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ACCESS_REVOKED", activitylog.ActivityLogEntryActionRemoved, activityLogEntryResourceTypeArtifactRegistryGithubRepository)
}

type ArtifactRegistryGithubRepositoryAccessGrantedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type ArtifactRegistryGithubRepositoryAccessRevokedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}
