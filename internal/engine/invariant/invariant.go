package invariant

import "github.com/base-org/pessimism/internal/core"

// ExecutionType ... Enum for execution type
type ExecutionType int

const (
	// HardCoded ... Hard coded execution type (ie native application code)
	HardCoded ExecutionType = iota
)

// Invariant ... Interface that all invariant implementations must adhere to
type Invariant interface {
	Addressing() bool
	UUID() core.InvSessionUUID
	WithUUID(sUUID core.InvSessionUUID)
	InputType() core.RegisterType
	Invalidate(core.TransitData) (*core.InvalOutcome, bool, error)
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
		inType:     inType,
		addressing: false,
	}

	for _, opt := range opts {
		opt(bi)
	}

	return bi
}

// UUID ... Returns the invariant session UUID
func (bi *BaseInvariant) UUID() core.InvSessionUUID {
	return bi.sUUID
}

// WithUUID ... Sets the invariant session UUID
func (bi *BaseInvariant) WithUUID(sUUID core.InvSessionUUID) {
	bi.sUUID = sUUID
}

// InputType ... Returns the input type for the invariant
func (bi *BaseInvariant) InputType() core.RegisterType {
	return bi.inType
}

// Invalidate ... Invalidates the invariant; defaults to no-op
func (bi *BaseInvariant) Invalidate(core.TransitData) (*core.InvalOutcome, bool, error) {
	return nil, false, nil
}

// Addressing ... Returns the boolean addressing property for the invariant
func (bi *BaseInvariant) Addressing() bool {
	return bi.addressing
}
