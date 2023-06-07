package invariant

import (
	"github.com/base-org/pessimism/internal/core"
)

// ExecutionType ... Enum for execution type
type ExecutionType int

const (
	// HardCoded ... Hard coded execution type (ie native application code)
	HardCoded ExecutionType = iota
)

// Invariant ... Interface that all invariant implementations must adhere to
type Invariant interface {
	InputType() core.RegisterType
	Invalidate(core.TransitData) (*core.InvalOutcome, bool, error)
	SUUID() core.InvSessionUUID
	SetSUUID(core.InvSessionUUID)
}

// BaseInvariantOpt ... Functional option for BaseInvariant
type BaseInvariantOpt = func(bi *BaseInvariant) *BaseInvariant

// WithAddressing ... Toggles addressing property for invariant
func WithAddressing() BaseInvariantOpt {
	return func(bi *BaseInvariant) *BaseInvariant {
		bi.addressing = true
		return bi
	}
}

// BaseInvariant ... Base invariant implementation
type BaseInvariant struct {
	addressing bool
	sUUID      core.InvSessionUUID
	inType     core.RegisterType
}

// NewBaseInvariant ... Initializer
func NewBaseInvariant(inType core.RegisterType,
	opts ...BaseInvariantOpt) Invariant {
	bi := &BaseInvariant{
		inType: inType,
	}

	for _, opt := range opts {
		opt(bi)
	}

	return bi
}

// SetSUUID ... Sets the invariant session UUID
func (bi *BaseInvariant) SetSUUID(sUUID core.InvSessionUUID) {
	bi.sUUID = sUUID
}

// SUUID ... Returns the invariant session UUID
func (bi *BaseInvariant) SUUID() core.InvSessionUUID {
	return bi.sUUID
}

// InputType ... Returns the input type for the invariant
func (bi *BaseInvariant) InputType() core.RegisterType {
	return bi.inType
}

// Invalidate ... Invalidates the invariant; defaults to no-op
func (bi *BaseInvariant) Invalidate(core.TransitData) (*core.InvalOutcome, bool, error) {
	return nil, false, nil
}
