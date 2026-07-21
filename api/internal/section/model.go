package section

import (
	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section/sectionsql"
)

//mgo:gen model
//mgo:gen order NAME CODE
//mgo:impl node paginated
type Section struct {
	Code      string     `json:"-"`
	Name      string     `json:"name"`
	ManagerId *uuid.UUID `json:"managerId"`
}

func (s Section) ID() ident.Ident {
	return NewIdent(s.Code)
}

func toGraphSection(s *sectionsql.Section) *Section {
	return &Section{
		Code:      s.Code,
		Name:      s.Name,
		ManagerId: s.ManagerID,
	}
}
