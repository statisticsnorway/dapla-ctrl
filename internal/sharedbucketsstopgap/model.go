package sharedbucketsstopgap

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-api/internal/graph/model"
	"github.com/statisticsnorway/dapla-api/internal/graph/pagination"
	"github.com/statisticsnorway/dapla-api/internal/sharedbucketsstopgap/sharedbucketsstopgapsql"
	"github.com/statisticsnorway/dapla-api/internal/slug"
)

type (
	SharedBucketConnection = pagination.Connection[*SharedBucket]
	SharedBucketEdge       = pagination.Edge[*SharedBucket]
)

type SharedBucket struct {
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	ShortName string    `json:"shortName"`
	Env       string    `json:"env"`
	TeamSlug  slug.Slug `json:"slug"`
}

func (SharedBucket) IsNode()       {}
func (SharedBucket) IsSearchNode() {}

func (s SharedBucket) ID() ident.Ident {
	return NewIdent(s.Name)
}

func toGraphBucket(s *sharedbucketsstopgapsql.SharedBucketsStopgap) *SharedBucket {
	return &SharedBucket{
		Name:      s.Name,
		Kind:      s.Kind,
		ShortName: s.ShortName,
		Env:       s.Env,
		TeamSlug:  s.TeamSlug,
	}
}

type SharedBucketOrder struct {
	Field     SharedBucketOrderField `json:"field"`
	Direction model.OrderDirection   `json:"direction"`
}

func (o *SharedBucketOrder) String() string {
	if o == nil {
		return ""
	}

	return strings.ToLower(o.Field.String() + ":" + o.Direction.String())
}

type SharedBucketOrderField string

const (
	SectionOrderFieldName SharedBucketOrderField = "NAME"
)

func (e SharedBucketOrderField) IsValid() bool {
	switch e {
	case SectionOrderFieldName:
		return true
	}
	return false
}

func (e SharedBucketOrderField) String() string {
	return string(e)
}

func (e *SharedBucketOrderField) UnmarshalGQL(v any) error {
	str, ok := v.(string)
	if !ok {
		return fmt.Errorf("enums must be strings")
	}

	*e = SharedBucketOrderField(str)
	if !e.IsValid() {
		return fmt.Errorf("%s is not a valid SharedBucketsOrderField", str)
	}
	return nil
}

func (e SharedBucketOrderField) MarshalGQL(w io.Writer) {
	fmt.Fprint(w, strconv.Quote(e.String()))
}
