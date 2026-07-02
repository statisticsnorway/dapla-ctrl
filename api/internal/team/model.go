package team

import (
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/model"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/team/teamsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/validate"
)

type (
	TeamConnection       = pagination.Connection[*Team]
	TeamEdge             = pagination.Edge[*Team]
	TeamMemberConnection = pagination.Connection[*TeamMember]
	TeamMemberEdge       = pagination.Edge[*TeamMember]
	TeamGroupConnection  = pagination.Connection[*TeamGroup]
	TeamGroupEdge        = pagination.Edge[*TeamGroup]
)

type Team struct {
	Slug                 slug.Slug  `json:"slug"`
	DisplayName          string     `json:"displayName"`
	SectionCode          string     `json:"sectionCode"`
	IsManaged            bool       `json:"isManaged"`
	HasManualEditing     bool       `json:"hasManualEditing"`
	LastSuccessfulSync   *time.Time `json:"lastSuccessfulSync"`
	DeleteKeyConfirmedAt *time.Time `json:"-"`
}

type ExternalReferences struct{}

func (Team) IsNode()           {}
func (Team) IsSearchNode()     {}
func (Team) IsActivityLogger() {}

func (t Team) DeletionInProgress() bool {
	return t.DeleteKeyConfirmedAt != nil
}

func (t Team) ID() ident.Ident {
	return newTeamIdent(t.Slug)
}

func (t *Team) ExternalResources() *TeamExternalResources {
	return &TeamExternalResources{
		team: t,
	}
}

type TeamOrder struct {
	Field     TeamOrderField       `json:"field"`
	Direction model.OrderDirection `json:"direction"`
}

func (o *TeamOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type TeamOrderField string

const (
	TeamOrderFieldSlug        TeamOrderField = "SLUG"
	TeamOrderFieldSectionCode TeamOrderField = "SECTION_CODE"
)

var AllTeamOrderField = []TeamOrderField{
	TeamOrderFieldSlug,
	TeamOrderFieldSectionCode,
}

func (e TeamOrderField) IsValid() bool {
	return slices.Contains(AllTeamOrderField, e)
}

func (e TeamOrderField) String() string {
	return string(e)
}

func (e *TeamOrderField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamOrderField", str)
	}
	return nil
}

func (e TeamOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func toGraphTeam(m *teamsql.Team) *Team {
	ret := &Team{
		Slug:             m.Slug,
		DisplayName:      m.DisplayName,
		SectionCode:      m.SectionCode,
		IsManaged:        m.IsManaged,
		HasManualEditing: m.HasManualEditing,
	}

	if m.LastSuccessfulSync.Valid {
		ret.LastSuccessfulSync = &m.LastSuccessfulSync.Time
	}

	if m.DeleteKeyConfirmedAt.Valid {
		ret.DeleteKeyConfirmedAt = &m.DeleteKeyConfirmedAt.Time
	}

	return ret
}

func toGraphTeamMember(teamSlug slug.Slug, userId uuid.UUID, groups []string) *TeamMember {
	return &TeamMember{
		TeamSlug: teamSlug,
		UserID:   userId,
		Groups:   groups,
	}
}

func toGraphUserTeam(m *teamsql.ListForUserRow) *TeamMember {
	return &TeamMember{
		TeamSlug: m.Slug,
		UserID:   m.UserID,
		Groups:   m.Groups,
	}
}

type TeamMember struct {
	TeamSlug slug.Slug `json:"-"`
	UserID   uuid.UUID `json:"-"`
	Groups   []string  `json:"-"`
}

type TeamAccessManager struct {
	TeamSlug slug.Slug `json:"-"`
	UserID   uuid.UUID `json:"-"`
}

type TeamMemberRole string

const (
	TeamMemberRoleMember TeamMemberRole = "MEMBER"
	TeamMemberRoleOwner  TeamMemberRole = "OWNER"
)

func (e TeamMemberRole) IsValid() bool {
	switch e {
	case TeamMemberRoleMember, TeamMemberRoleOwner:
		return true
	}
	return false
}

func (e TeamMemberRole) String() string {
	return string(e)
}

func (e *TeamMemberRole) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamMemberRole(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamMemberRole", str)
	}
	return nil
}

func (e TeamMemberRole) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type TeamMemberOrder struct {
	Field     TeamMemberOrderField `json:"field"`
	Direction model.OrderDirection `json:"direction"`
}

func (o *TeamMemberOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type TeamMemberOrderField string

const (
	TeamMemberOrderFieldName  TeamMemberOrderField = "NAME"
	TeamMemberOrderFieldEmail TeamMemberOrderField = "EMAIL"
)

func (e TeamMemberOrderField) IsValid() bool {
	switch e {
	case TeamMemberOrderFieldName, TeamMemberOrderFieldEmail:
		return true
	}
	return false
}

func (e TeamMemberOrderField) String() string {
	return string(e)
}

func (e *TeamMemberOrderField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = TeamMemberOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid TeamMemberOrderField", str)
	}
	return nil
}

func (e TeamMemberOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

func toGraphTeamDeleteKey(key *teamsql.TeamDeleteKey) *TeamDeleteKey {
	var confirmedAt *time.Time
	if key.ConfirmedAt.Valid {
		confirmedAt = &key.ConfirmedAt.Time
	}
	return &TeamDeleteKey{
		KeyUUID:         key.Key,
		CreatedAt:       key.CreatedAt.Time,
		ConfirmedAt:     confirmedAt,
		CreatedByUserID: key.CreatedBy,
		TeamSlug:        key.TeamSlug,
	}
}

type UserTeamOrder struct {
	Field     UserTeamOrderField   `json:"field"`
	Direction model.OrderDirection `json:"direction"`
}

func (o *UserTeamOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type UserTeamOrderField string

const (
	UserTeamOrderFieldTeamSlug UserTeamOrderField = "TEAM_SLUG"
)

func (e UserTeamOrderField) IsValid() bool {
	switch e {
	case UserTeamOrderFieldTeamSlug:
		return true
	}
	return false
}

func (e UserTeamOrderField) String() string {
	return string(e)
}

func (e *UserTeamOrderField) UnmarshalGQL(v interface{}) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = UserTeamOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid UserTeamOrderField", str)
	}
	return nil
}

func (e UserTeamOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}

type CreateTeamInput struct {
	Slug        slug.Slug `json:"slug"`
	DisplayName string    `json:"displayName"`
	SectionCode string    `json:"sectionCode"`
	IsManaged   *bool     `json:"isManaged"`
}

func (i *CreateTeamInput) Validate(ctx context.Context) error {
	verr := validate.New()
	i.DisplayName = strings.TrimSpace(i.DisplayName)

	if slices.ContainsFunc([]string{"managers", "consumers", "data-admins", "developers"}, func(category string) bool {
		return strings.Contains(i.Slug.String(), category)
	}) {
		verr.Add("slug", "Team slug cannot contain a group category.")
	}

	if available, err := db(ctx).SlugAvailable(ctx, i.Slug); err != nil {
		return err
	} else if !available {
		verr.Add("slug", "Team slug is not available.")
	}

	if i.DisplayName == "" {
		verr.Add("displayName", "This is not a valid display name.")
	}

	if i.IsManaged == nil {
		verr.Add("isManaged", "Whether the team should be managed must be specified.")
	}

	return verr.NilIfEmpty()
}

type UpdateTeamInput struct {
	Slug             slug.Slug `json:"slug"`
	DisplayName      *string   `json:"displayName"`
	SectionCode      *string   `json:"sectionCode"`
	HasManualEditing *bool     `json:"hasManualEditing"`
}

func (i *UpdateTeamInput) Validate() error {
	verr := validate.New()

	if i.DisplayName != nil {
		i.DisplayName = new(strings.TrimSpace(*i.DisplayName))
	}

	if i.DisplayName != nil && *i.DisplayName == "" {
		verr.Add("displayName", "This is not a valid display name.")
	}

	return verr.NilIfEmpty()
}

func (i *UpdateTeamInput) HasNoChanges() bool {
	return i.DisplayName == nil && i.SectionCode == nil && i.HasManualEditing == nil
}

type AddTeamAccessManagerInput struct {
	TeamSlug  slug.Slug `json:"teamSlug"`
	UserEmail string    `json:"userEmail"`
	UserId    uuid.UUID `json:"-"`
}

type RemoveTeamAccessManagerInput struct {
	TeamSlug  slug.Slug `json:"teamSlug"`
	UserEmail string    `json:"userEmail"`
	UserId    uuid.UUID `json:"-"`
}

type CreateTeamPayload struct {
	Team *Team `json:"team"`
}

type UpdateTeamPayload struct {
	Team *Team `json:"team"`
}

type AddTeamAccessManagerPayload struct {
	TeamSlug slug.Slug `json:"-"`
	UserId   uuid.UUID `json:"-"`
}

type RemoveTeamAccessManagerPayload struct {
	TeamSlug slug.Slug `json:"-"`
	UserId   uuid.UUID `json:"-"`
}

type RequestTeamDeletionInput struct {
	Slug slug.Slug `json:"slug"`
}

type RequestTeamDeletionPayload struct {
	Key *TeamDeleteKey `json:"key"`
}

type TeamDeleteKey struct {
	KeyUUID         uuid.UUID  `json:"key"`
	CreatedAt       time.Time  `json:"createdAt"`
	ConfirmedAt     *time.Time `json:"-"`
	CreatedByUserID uuid.UUID  `json:"-"`
	TeamSlug        slug.Slug  `json:"-"`
}

func (t TeamDeleteKey) Key() string {
	return t.KeyUUID.String()
}

func (t *TeamDeleteKey) Expires() time.Time {
	return t.CreatedAt.Add(time.Hour)
}

func (t *TeamDeleteKey) HasExpired() bool {
	return time.Now().After(t.Expires())
}

type ConfirmTeamDeletionInput struct {
	Key  string    `json:"key"`
	Slug slug.Slug `json:"slug"`
}

type ConfirmTeamDeletionPayload struct {
	DeletionStarted bool `json:"deletionStarted"`
}

type AddTeamMemberInput struct {
	TeamSlug  slug.Slug      `json:"teamSlug"`
	UserEmail string         `json:"userEmail"`
	Role      TeamMemberRole `json:"role"`
	UserID    uuid.UUID      `json:"-"`
}

type AddTeamMemberPayload struct {
	Member *TeamMember `json:"member"`
}

type RemoveTeamMemberInput struct {
	TeamSlug  slug.Slug `json:"teamSlug"`
	UserEmail string    `json:"userEmail"`
	UserID    uuid.UUID `json:"-"`
}

type RemoveTeamMemberPayload struct {
	UserID   uuid.UUID `json:"-"`
	TeamSlug slug.Slug `json:"-"`
}

type SetTeamMemberRoleInput struct {
	TeamSlug  slug.Slug      `json:"teamSlug"`
	UserEmail string         `json:"userEmail"`
	Role      TeamMemberRole `json:"role"`
	UserID    uuid.UUID      `json:"-"`
}

type SetTeamMemberRolePayload struct {
	Member *TeamMember `json:"member"`
}

type TeamInventoryCounts struct {
	TeamSlug slug.Slug `json:"-"`
}

type TeamExternalResources struct {
	team *Team
}

type TeamGroup struct {
	TeamSlug  slug.Slug `json:"teamSlug"`
	GroupName string    `json:"groupName"`
}
