package sharedbucketsstopgap

import (
	"strings"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/sharedbucketsstopgap/sharedbucketsstopgapsql"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

//mgo:gen model
//mgo:gen order NAME KIND SHORT_NAME ENV TEAM
//mgo:impl node searchnode paginated
type SharedBucket struct {
	Name      string    `json:"name"`
	Kind      string    `json:"kind"`
	ShortName string    `json:"shortName"`
	Env       string    `json:"env"`
	TeamSlug  slug.Slug `json:"slug"`
}

func (s SharedBucket) ID() ident.Ident {
	return NewIdent(s.Name)
}

//mgo:gen model
//mgo:impl paginated
type SharedBucketAccess struct {
	BucketName string    `json:"-"`
	TeamSlug   slug.Slug `json:"-"`
	GroupNames []string  `json:"-"`
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

type SharedBucketFilter struct {
	Envs  []string `json:"envs,omitempty"`
	Kinds []string `json:"kinds,omitempty"`
}

func (f *SharedBucketFilter) EnvFilter() []string {
	var envFilter []string
	if f != nil {
		for _, c := range f.Envs {
			envFilter = append(envFilter, strings.ToLower(c))
		}
	}
	return envFilter
}

func (f *SharedBucketFilter) KindFilter() []string {
	var kindFilter []string
	if f != nil {
		for _, c := range f.Kinds {
			kindFilter = append(kindFilter, strings.ToLower(c))
		}
	}
	return kindFilter
}
