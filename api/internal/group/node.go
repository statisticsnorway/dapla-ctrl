package group

import (
	"fmt"

	"github.com/statisticsnorway/dapla-api/internal/graph/ident"
)

// identType is an enumeration type for the different identifiers defined by this domain package. Each domain package
// must have its own enumeration type to not cause collisions with other domain packages.
type identType int

const (
	// A list of identifiers defined by this package.

	identKey identType = iota
)

func init() {
	// Register all identifiers during initialization. The first argument is the constant identifier type, the second is
	// a globally unique string representation of the identifier type, and should be as short as possible. The third
	// argument is a lookup function that can be used to retrieve the node associated with the identifier. The return
	// type of the lookup function must be compatible with the model.Node interface.
	//
	// Refer to https://go.dev/doc/effective_go#init for more information about the init() function itself.

	ident.RegisterIdentType(identKey, "G", GetByIdent)
}

func NewIdent(groupName string) ident.Ident {
	return ident.NewIdent(identKey, groupName)
}

func parseIdent(id ident.Ident) (string, error) {
	parts := id.Parts()
	if len(parts) != 1 {
		return "", fmt.Errorf("invalid group ident")
	}

	return parts[0], nil
}
