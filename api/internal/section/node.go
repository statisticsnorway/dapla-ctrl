package section

import (
	"fmt"

	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
)

type identType int

const (
	identKey identType = iota
)

func init() {
	ident.RegisterIdentType(identKey, "S", GetByIdent)
}

func NewIdent(code string) ident.Ident {
	return ident.NewIdent(identKey, code)
}

func parseIdent(id ident.Ident) (string, error) {
	parts := id.Parts()
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid section ident")
	}

	return parts[0], nil
}
