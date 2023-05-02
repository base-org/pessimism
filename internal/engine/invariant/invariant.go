package invariant

import "github.com/base-org/pessimism/internal/core"

type InvariantType int

const (
	HardCoded InvariantType = iota
)

type Invariant interface {
	InputType() core.RegisterType
	Invalidate(core.TransitData) (bool, error)
}
