package section

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/model"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/section/sectionsql"
)

type (
	SectionConnection = pagination.Connection[*Section]
	SectionEdge       = pagination.Edge[*Section]
)

type Section struct {
	Code      string     `json:"-"`
	Name      string     `json:"name"`
	ManagerId *uuid.UUID `json:"managerId"`
}

func (Section) IsNode() {}

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

type SectionOrder struct {
	Field     SectionOrderField    `json:"field"`
	Direction model.OrderDirection `json:"direction"`
}

func (o *SectionOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type SectionOrderField string

const (
	SectionOrderFieldName SectionOrderField = "NAME"
	SectionOrderFieldCode SectionOrderField = "CODE"
)

func (e SectionOrderField) IsValid() bool {
	switch e {
	case SectionOrderFieldName, SectionOrderFieldCode:
		return true
	}
	return false
}

func (e SectionOrderField) String() string {
	return string(e)
}

func (e *SectionOrderField) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SectionOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SectionOrderField", str)
	}
	return nil
}

func (e SectionOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
