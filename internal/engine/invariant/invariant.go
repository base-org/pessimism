//go:generate mockgen -package mocks --destination ../../mocks/invariant.go . Invariant

package invariant

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// ExecutionType ... Enum for execution type
type ExecutionType int

const (
	// HardCoded ... Hard coded execution type (ie native application code)
	HardCoded ExecutionType = iota

	invalidInTypeErr = "invalid input type provided for invariant. expected %s, got %s"
)

// Invariant ... Interface that all invariant implementations must adhere to
type Invariant interface {
	InputType() core.RegisterType
	ValidateInput(core.TransitData) error
	Invalidate(core.TransitData) (*core.Invalidation, bool, error)
	SUUID() core.SUUID
	SetSUUID(core.SUUID)
}

// BaseInvariantOpt ... Functional option for BaseInvariant
type BaseInvariantOpt = func(bi *BaseInvariant) *BaseInvariant

// BaseInvariant ... Base invariant implementation
type BaseInvariant struct {
	sUUID  core.SUUID
	inType core.RegisterType
}

// NewBaseInvariant ... Initializer for BaseInvariant
// This is a base type that's inherited by all hardcoded
// invariant implementations
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

// SUUID ... Returns the invariant session UUID
func (bi *BaseInvariant) SUUID() core.SUUID {
	return bi.sUUID
}

// InputType ... Returns the input type for the invariant
func (bi *BaseInvariant) InputType() core.RegisterType {
	return bi.inType
}

// Invalidate ... Invalidates the invariant; defaults to no-op
func (bi *BaseInvariant) Invalidate(core.TransitData) (*core.Invalidation, bool, error) {
	return nil, false, nil
}

// SetSUUID ... Sets the invariant session UUID
func (bi *BaseInvariant) SetSUUID(sUUID core.SUUID) {
	bi.sUUID = sUUID
}

func (bi *BaseInvariant) ValidateInput(td core.TransitData) error {
	if td.Type != bi.InputType() {
		return fmt.Errorf(invalidInTypeErr, bi.InputType(), td.Type)
	}

	return nil
}
