package artifactregistry

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/slug"
)

type identType int

const (
	identKeyGHRepo identType = iota
	identKeyARRepo
)

func init() {
	ident.RegisterIdentType(identKeyGHRepo, "ARGHR", getGHReposByIdent)
	ident.RegisterIdentType(identKeyARRepo, "ARR", getARRepoByIdent)
}

func newARIdent(teamSlug slug.Slug, format string) ident.Ident {
	return ident.NewIdent(identKeyARRepo, teamSlug.String(), format)
}

func newGHIdent(teamSlug slug.Slug, githubRepositoryName string) ident.Ident {
	return ident.NewIdent(identKeyGHRepo, teamSlug.String(), githubRepositoryName)
}

func parseARIdent(id ident.Ident) (slug.Slug, ArtifactRegistryFormat, error) {
	parts := id.Parts()
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository ident")
	}

	format := ArtifactRegistryFormat(parts[1])
	if !format.IsValid() {
		return "", "", fmt.Errorf("invalid AR format %q", format.String())
	}

	return slug.Slug(parts[0]), format, nil
}

func parseGHIdent(id ident.Ident) (slug.Slug, string, error) {
	parts := id.Parts()
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repository ident")
	}

	return slug.Slug(parts[0]), parts[1], nil
}
