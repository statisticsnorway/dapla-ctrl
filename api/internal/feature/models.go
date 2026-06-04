package feature

import "github.com/statisticsnorway/dapla-ctrl/api/internal/graph/ident"

type Features struct{}

func (f Features) ID() ident.Ident { return NewIdent("container") }
func (f Features) IsNode()         {}
