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
	ident.RegisterIdentType(identKey, "AR", getByIdent)
}

func newIdent(teamSlug slug.Slug, format string) ident.Ident {
	return ident.NewIdent(identKey, teamSlug.String(), format)
}

func parseIdent(id ident.Ident) (slug.Slug, string, error) {
	parts := id.Parts()
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository ident")
	}

	return slug.Slug(parts[0]), parts[1], nil
}
