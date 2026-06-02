package group

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-api/internal/activitylog"
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/user"
)

const (
	ActivityLogEntryResourceTypeGroup activitylog.ActivityLogEntryResourceType = "GROUP"
)

func init() {
	activitylog.RegisterTransformer(ActivityLogEntryResourceTypeGroup, func(entry activitylog.GenericActivityLogEntry) (activitylog.ActivityLogEntry, error) {
		switch entry.Action {
		case activitylog.ActivityLogEntryActionCreated:
			return GroupCreatedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Created group"),
			}, nil
		case activitylog.ActivityLogEntryActionAdded:
			data, err := activitylog.TransformData(entry, func(data *GroupMemberAddedActivityLogEntryData) *GroupMemberAddedActivityLogEntryData {
				return data
			})
			if err != nil {
				return nil, err
			}
			return GroupMemberAddedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Add member"),
				Data:                    data,
			}, nil
		case activitylog.ActivityLogEntryActionRemoved:
			data, err := activitylog.TransformData(entry, func(data *GroupMemberRemovedActivityLogEntryData) *GroupMemberRemovedActivityLogEntryData {
				return data
			})
			if err != nil {
				return nil, err
			}
			return GroupMemberRemovedActivityLogEntry{
				GenericActivityLogEntry: entry.WithMessage("Remove member"),
				Data:                    data,
			}, nil
		default:
			return nil, fmt.Errorf("unsupported team activity log entry action: %q", entry.Action)
		}
	})

	activitylog.RegisterFilter("GROUP_CREATED", activitylog.ActivityLogEntryActionCreated, ActivityLogEntryResourceTypeGroup)
	activitylog.RegisterFilter("GROUP_MEMBER_ADDED", activitylog.ActivityLogEntryActionAdded, ActivityLogEntryResourceTypeGroup)
	activitylog.RegisterFilter("GROUP_MEMBER_REMOVED", activitylog.ActivityLogEntryActionRemoved, ActivityLogEntryResourceTypeGroup)
}

type GroupCreatedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
}

type GroupMemberAddedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
	Data *GroupMemberAddedActivityLogEntryData `json:"data"`
}

type GroupMemberAddedActivityLogEntryData struct {
	UserUUID  uuid.UUID `json:"userID"`
	UserEmail string    `json:"userEmail"`
}

func (t GroupMemberAddedActivityLogEntryData) UserID() ident.Ident {
	return user.NewIdent(t.UserUUID)
}

type GroupMemberRemovedActivityLogEntry struct {
	activitylog.GenericActivityLogEntry
	Data *GroupMemberRemovedActivityLogEntryData `json:"data"`
}

type GroupMemberRemovedActivityLogEntryData struct {
	UserUUID  uuid.UUID `json:"userID"`
	UserEmail string    `json:"userEmail"`
}

func (t GroupMemberRemovedActivityLogEntryData) UserID() ident.Ident {
	return user.NewIdent(t.UserUUID)
}
