package invariant

import "github.com/base-org/pessimism/internal/core"

type ExecutionType int

const (
	HardCoded ExecutionType = iota
)

type Invariant interface {
	InputType() core.RegisterType
	Invalidate(core.TransitData) (bool, error)
}
