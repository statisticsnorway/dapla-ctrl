package message

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"
)

// identType is an enumeration type for the different identifiers defined by this domain package. Each domain package
// must have its own enumeration type to not cause collisions with other domain packages.
type identType int

const (
	// A list of identifiers defined by this package.

	identMessage identType = iota
	identMessageEnvironment
)

func init() {
	// Register all identifiers during initialization. The first argument is the constant identifier type, the second is
	// a globally unique string representation of the identifier type, and should be as short as possible. The third
	// argument is a lookup function that can be used to retrieve the node associated with the identifier. The return
	// type of the lookup function must be compatible with the model.Node interface.
	//
	// Refer to https://go.dev/doc/effective_go#init for more information about the init() function itself.

	ident.RegisterIdentType(identMessage, "M", GetByIdent)
}

// newMessageIdent creates a new identifier for a specific message
func newMessageIdent(messageId uuid.UUID) ident.Ident {
	return ident.NewIdent(identMessage, messageId.String())
}

// parseMessageIdent returns the message uuid from a message identifier. If the identifier is invalid, an error is returned.
func parseMessageIdent(id ident.Ident) (uuid.UUID, error) {
	parts := id.Parts()
	if len(parts) != 1 {
		return uuid.Nil, fmt.Errorf("invalid message ident")
	}

	return uuid.Parse(parts[0])
}
