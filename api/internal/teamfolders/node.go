package teamfolders

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
	ident.RegisterIdentType(identKey, "TF", GetByIdent)
}

func NewIdent(teamSlug slug.Slug, env string) ident.Ident {
	return ident.NewIdent(identKey, teamSlug.String(), env)
}

func parseIdent(id ident.Ident) (slug.Slug, string, error) {
	parts := id.Parts()
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid team folder ident")
	}

	return slug.Slug(parts[0]), parts[1], nil
}
