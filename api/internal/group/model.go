package group

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/model"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/group/groupsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/validate"
)

type (
	GroupConnection       = pagination.Connection[*Group]
	GroupEdge             = pagination.Edge[*Group]
	GroupMemberConnection = pagination.Connection[*GroupMember]
	GroupMemberEdge       = pagination.Edge[*GroupMember]
)

type Group struct {
	Name     string    `json:"name"`
	TeamSlug slug.Slug `json:"slug"`
	Category string    `json:"category"`
	Suffix   string    `json:"suffix"`
}

type ExternalReferences struct{}

func (Group) IsNode()           {}
func (Group) IsSearchNode()     {}
func (Group) IsActivityLogger() {}

func (g Group) ID() ident.Ident {
	return NewIdent(g.Name)
}

type GroupOrder struct {
	Field     GroupOrderField      `json:"field"`
	Direction model.OrderDirection `json:"direction"`
}

func (o *GroupOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type GroupOrderField string

const (
	GroupOrderFieldName     GroupOrderField = "NAME"
	GroupOrderFieldSlug     GroupOrderField = "TEAM"
	GroupOrderFieldCategory GroupOrderField = "CATEGORY"
)

var AllGroupOrderField = []GroupOrderField{
	GroupOrderFieldName,
	GroupOrderFieldSlug,
	GroupOrderFieldCategory,
}

func (e GroupOrderField) IsValid() bool {
	return slices.Contains(AllGroupOrderField, e)
}

func (e GroupOrderField) String() string {
	return string(e)
}

func (e *GroupOrderField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = GroupOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid GroupOrderField", str)
	}
	return nil
}

func (e GroupOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func toGraphGroup(m *groupsql.Group) *Group {
	ret := &Group{
		Name:     m.Name,
		TeamSlug: m.TeamSlug,
		Category: m.Category,
		Suffix:   m.Suffix,
	}
	return ret
}

func toGraphGroupMember(m *groupsql.ListMembersRow) *GroupMember {
	return &GroupMember{
		GroupName: m.Group.Name,
		UserID:    m.User.ID,
	}
}

func toGraphUserGroup(m *groupsql.ListForUserRow) *GroupMember {
	return &GroupMember{
		GroupName: m.GroupName,
		UserID:    m.UserID,
	}
}

type GroupMember struct {
	GroupName string    `json:"-"`
	UserID    uuid.UUID `json:"-"`
}

type GroupFilter struct {
	Categories []string `json:"categories,omitempty"`
}

func (f *GroupFilter) CategoryFilter() []string {
	var categoryFilter []string
	if f != nil {
		for _, c := range f.Categories {
			categoryFilter = append(categoryFilter, strings.ToLower(c))
		}
	}
	return categoryFilter
}

type CreateGroupInput struct {
	TeamSlug slug.Slug `json:"teamSlug"`
	Category string    `json:"category"`
	Suffix   *string   `json:"suffix,omitempty"`
}

func (i *CreateGroupInput) Validate(ctx context.Context) error {
	verr := validate.New()

	// check if team exists
	if exists, err := db(ctx).TeamExists(ctx, i.TeamSlug); err != nil {
		return err
	} else if !exists {
		verr.Add("teamSlug", "Team with the given slug does not exists.")
	}

	if !slices.Contains([]string{"managers", "consumers", "data-admins", "developers"}, i.Category) {
		verr.Add("category", "Invalid category provided")
	}

	if i.Suffix == nil {
		i.Suffix = new("")
	}

	if exists, err := db(ctx).GroupExists(ctx, groupsql.GroupExistsParams{
		TeamSlug: i.TeamSlug,
		Category: i.Category,
		Suffix:   *i.Suffix,
	}); err != nil {
		return err
	} else if exists {
		verr.Add("teamSlug", "Group with the same team, category and suffix already exists.")
	}

	return verr.NilIfEmpty()
}

type CreateGroupPayload struct {
	Group *Group `json:"group"`
}

type AddGroupMemberInput struct {
	GroupName string    `json:"groupName"`
	UserEmail string    `json:"userEmail"`
	UserID    uuid.UUID `json:"-"`
}

type AddGroupMemberPayload struct {
	Member *GroupMember `json:"member"`
}

type RemoveGroupMemberInput struct {
	GroupName string    `json:"groupName"`
	UserEmail string    `json:"userEmail"`
	UserID    uuid.UUID `json:"-"`
}

type RemoveGroupMemberPayload struct {
	GroupName string    `json:"-"`
	UserID    uuid.UUID `json:"-"`
}
