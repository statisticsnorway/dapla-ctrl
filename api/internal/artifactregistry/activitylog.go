package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
)

const (
	activityLogEntryResourceTypeArtifactRegistryGithubRepository activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_GITHUB_REPOSITORY"
)

func init() {
	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryGithubRepository, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionAdded:
			return ArtifactRegistryGithubRepositoryAddedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Added artifact registry github repository to team"),
			}, nil
		case activitylog.ActivityLogEntryActionRemoved:
			return ArtifactRegistryGithubRepositoryRemovedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Removed artifact registry github repository from team"),
			}, nil

		default:
			return nil, fmt.Errorf("unsupported repository activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_ADDED", activitylog.ActivityLogEntryActionAdded, activityLogEntryResourceTypeArtifactRegistryGithubRepository)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_GITHUB_REPOSITORY_REMOVED", activitylog.ActivityLogEntryActionRemoved, activityLogEntryResourceTypeArtifactRegistryGithubRepository)
}

type ArtifactRegistryGithubRepositoryAddedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type ArtifactRegistryGithubRepositoryRemovedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}
