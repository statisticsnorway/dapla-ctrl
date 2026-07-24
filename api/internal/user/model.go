package user

import (
	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/user/usersql"
)

//mgo:gen model
//mgo:gen order NAME EMAIL SECTION_CODE
//mgo:impl node searchnode paginated
type User struct {
	UUID           uuid.UUID `json:"-"`
	Email          string    `json:"email"`
	Name           string    `json:"name"`
	JobTitle       *string   `json:"jobTitle"`
	ExternalID     string    `json:"externalId"`
	Admin          bool      `json:"admin"`
	SectionCode    *string   `json:"sectionCode"`
	EmploymentType string    `json:"employmentType"`
}

func (u *User) GetID() uuid.UUID       { return u.UUID }
func (u *User) Identity() string       { return u.Email }
func (u *User) IsServiceAccount() bool { return false }
func (u *User) IsAdmin() bool          { return u.Admin }

func (u User) ID() ident.Ident {
	return NewIdent(u.UUID)
}

func toGraphUser(u *usersql.User) *User {
	return &User{
		UUID:           u.ID,
		Email:          u.Email,
		Name:           u.Name,
		JobTitle:       u.JobTitle,
		ExternalID:     u.ExternalID,
		Admin:          u.Admin,
		SectionCode:    u.SectionCode,
		EmploymentType: u.EmploymentType,
	}
}

type AuthenticatedUser interface {
	GetID() uuid.UUID
	Identity() string
	IsServiceAccount() bool
}
