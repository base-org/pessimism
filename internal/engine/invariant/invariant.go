package invariant

import "github.com/base-org/pessimism/internal/core"

type ExecutionType int

const (
	HardCoded ExecutionType = iota
)

type Invariant interface {
	Addressing() bool
	UUID() core.InvSessionUUID
	WithUUID(sUUID core.InvSessionUUID)
	InputType() core.RegisterType
	Invalidate(core.TransitData) (bool, error)
}

type BaseInvariantOpt = func(bi *BaseInvariant) *BaseInvariant

func WithAddressing() BaseInvariantOpt {
	return func(bi *BaseInvariant) *BaseInvariant {
		bi.addresing = true
		return bi
	}
}

type BaseInvariant struct {
	addresing bool
	sUUID     core.InvSessionUUID
	inType    core.RegisterType
}

func NewBaseInvariant(inType core.RegisterType,
	opts ...BaseInvariantOpt) Invariant {
	bi := &BaseInvariant{
		inType:    inType,
		addresing: false,
	}

	for _, opt := range opts {
		opt(bi)
	}

	return bi
}
func (bi *BaseInvariant) UUID() core.InvSessionUUID {
	return bi.sUUID
}

func (bi *BaseInvariant) WithUUID(sUUID core.InvSessionUUID) {
	bi.sUUID = sUUID
}

func (bi *BaseInvariant) InputType() core.RegisterType {
	return bi.inType
}

func (bi *BaseInvariant) Invalidate(core.TransitData) (bool, error) {
	return false, nil
}

func (bi *BaseInvariant) Addressing() bool {
	return bi.addresing
}
