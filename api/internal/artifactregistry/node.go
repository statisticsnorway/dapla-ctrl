package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

type identType int

const (
	identKey identType = iota
)

func init() {
	ident.RegisterIdentType(identKey, "ARGHR", getByIdent)
}

func newIdent(teamSlug slug.Slug, githubRepositoryName string) ident.Ident {
	return ident.NewIdent(identKey, teamSlug.String(), githubRepositoryName)
}

func parseIdent(id ident.Ident) (slug.Slug, string, error) {
	parts := id.Parts()
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository ident")
	}

	return slug.Slug(parts[0]), parts[1], nil
}
