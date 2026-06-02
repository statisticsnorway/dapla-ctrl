package sharedbucketsstopgap

import (
	"fmt"

	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
)

type identType int

const (
	identKey identType = iota
)

func init() {
	ident.RegisterIdentType(identKey, "SBS", GetByIdent)
}

func NewIdent(code string) ident.Ident {
	return ident.NewIdent(identKey, code)
}

func parseIdent(id ident.Ident) (string, error) {
	parts := id.Parts()
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid bucket ident")
	}

	return parts[0], nil
}
