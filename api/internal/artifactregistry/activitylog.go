package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/activitylog"
)

const (
	activityLogEntryResourceTypeArtifactRegistryRepository activitylog.ActivityLogEntryResourceType = "ARTIFACT_REGISTRY_REPOSITORY"
)

func init() {
	activitylog.RegisterTransformer(activityLogEntryResourceTypeArtifactRegistryRepository, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionAdded:
			return TeamArtifactRegistryRepositoryAddedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Added repository to team"),
			}, nil
		case activitylog.ActivityLogEntryActionRemoved:
			return TeamArtifactRegistryRepositoryRemovedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Removed repository from team"),
			}, nil

		default:
			return nil, fmt.Errorf("unsupported repository activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterFilter("ARTIFACT_REGISTRY_REPOSITORY_ADDED", activitylog.ActivityLogEntryActionAdded, activityLogEntryResourceTypeArtifactRegistryRepository)
	activitylog.RegisterFilter("ARTIFACT_REGISTRY_REPOSITORY_REMOVED", activitylog.ActivityLogEntryActionRemoved, activityLogEntryResourceTypeArtifactRegistryRepository)
}

type TeamArtifactRegistryRepositoryAddedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type TeamArtifactRegistryRepositoryRemovedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}
