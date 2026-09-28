package teambuckets

import (
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/teambuckets/teambucketssql"
)

//mgo:gen model
//mgo:gen order NAME KIND ENV TEAM
//mgo:impl node searchnode paginated
type TeamBucket struct {
	Name     string    `json:"name"`
	Kind     string    `json:"kind"`
	Env      string    `json:"env"`
	TeamSlug slug.Slug `json:"slug"`
}

func (s TeamBucket) ID() ident.Ident {
	return NewIdent(s.Name)
}

func toGraphBucket(s *teambucketssql.GetByNamesRow) *TeamBucket {
	return &TeamBucket{
		Name:     s.Name,
		Kind:     s.Kind,
		Env:      s.Env,
		TeamSlug: s.TeamSlug,
	}
}

type TeamBucketFilter struct {
	Envs  []string `json:"envs,omitempty"`
	Kinds []string `json:"kinds,omitempty"`
}

func (f *TeamBucketFilter) EnvFilter() []string {
	var envFilter []string
	if f != nil {
		for _, c := range f.Envs {
			envFilter = append(envFilter, strings.ToLower(c))
		}
	}
	return envFilter
}

func (f *TeamBucketFilter) KindFilter() []string {
	var kindFilter []string
	if f != nil {
		for _, c := range f.Kinds {
			kindFilter = append(kindFilter, strings.ToLower(c))
		}
	}
	return kindFilter
}
