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
	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/model"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/slug"
	"github.com/statisticsnorway/dapla-api/internal/team/teamsql"
	"github.com/statisticsnorway/dapla-api/internal/validate"
	"k8s.io/utils/ptr"
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
	Purpose              string     `json:"purpose"`
	SectionCode          string     `json:"sectionCode"`
	LastSuccessfulSync   *time.Time `json:"lastSuccessfulSync"`
	DeleteKeyConfirmedAt *time.Time `json:"-"`
}

type ExternalReferences struct{}

func (Team) IsNode()       {}
func (Team) IsSearchNode() {}

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
	TeamOrderFieldSlug TeamOrderField = "SLUG"
)

var AllTeamOrderField = []TeamOrderField{
	TeamOrderFieldSlug,
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
		Slug:        m.Slug,
		Purpose:     m.Purpose,
		SectionCode: m.SectionCode,
	}

	if m.LastSuccessfulSync.Valid {
		ret.LastSuccessfulSync = &m.LastSuccessfulSync.Time
	}

	if m.DeleteKeyConfirmedAt.Valid {
		ret.DeleteKeyConfirmedAt = &m.DeleteKeyConfirmedAt.Time
	}

	return ret
}

func toGraphTeamMember(m *teamsql.ListMembersRow) *TeamMember {
	return &TeamMember{
		TeamSlug: m.TeamSlug,
		UserID:   m.ID,
		Groups:   m.Groups,
	}
}

func toGraphUserTeam(m *teamsql.ListForUserRow) *TeamMember {
	return &TeamMember{
		TeamSlug: m.Slug,
		UserID:   m.ID,
		Groups:   m.Groups,
	}
}

type TeamMember struct {
	TeamSlug slug.Slug `json:"-"`
	UserID   uuid.UUID `json:"-"`
	Groups   []string  `json:"-"`
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
	TeamMemberOrderFieldRole  TeamMemberOrderField = "ROLE"
)

func (e TeamMemberOrderField) IsValid() bool {
	switch e {
	case TeamMemberOrderFieldName, TeamMemberOrderFieldEmail, TeamMemberOrderFieldRole:
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
	Purpose     string    `json:"purpose"`
	SectionCode string    `json:"sectionCode"`
}

func (i *CreateTeamInput) Validate(ctx context.Context) error {
	verr := validate.New()
	i.Purpose = strings.TrimSpace(i.Purpose)

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

	if i.Purpose == "" {
		verr.Add("purpose", "This is not a valid purpose.")
	}

	return verr.NilIfEmpty()
}

type UpdateTeamInput struct {
	Slug    slug.Slug `json:"slug"`
	Purpose *string   `json:"purpose" `
}

func (i *UpdateTeamInput) Validate() error {
	verr := validate.New()

	if i.Purpose != nil {
		i.Purpose = ptr.To(strings.TrimSpace(*i.Purpose))
	}

	if i.Purpose != nil && *i.Purpose == "" {
		verr.Add("purpose", "This is not a valid purpose.")
	}

	return verr.NilIfEmpty()
}

type CreateTeamPayload struct {
	Team *Team `json:"team"`
}

type UpdateTeamPayload struct {
	Team *Team `json:"team"`
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
