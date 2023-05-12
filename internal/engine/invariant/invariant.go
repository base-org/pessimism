package invariant

import "github.com/base-org/pessimism/internal/core"

type ExecutionType int

const (
	HardCoded ExecutionType = iota
)

type Invariant interface {
	UUID() core.InvSessionUUID
	WithUUID(sUUID core.InvSessionUUID)
	InputType() core.RegisterType
	Invalidate(core.TransitData) (bool, error)
}

type BaseInvariant struct {
	sUUID  core.InvSessionUUID
	inType core.RegisterType
}

func NewBaseInvariant(inType core.RegisterType) Invariant {
	return &BaseInvariant{
		inType: inType,
	}
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
