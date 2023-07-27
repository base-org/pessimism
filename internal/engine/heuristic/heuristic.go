//go:generate mockgen -package mocks --destination ../../mocks/heuristic.go . Heuristic

package heuristic

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

// ExecutionType ... Enum for execution type
type ExecutionType int

const (
	// HardCoded ... Hard coded execution type (ie native application code)
	HardCoded ExecutionType = iota

	invalidInTypeErr = "invalid input type provided for heuristic. expected %s, got %s"
)

// Heuristic ... Interface that all heuristic implementations must adhere to
type Heuristic interface {
	InputType() core.RegisterType
	ValidateInput(core.TransitData) error
	Invalidate(core.TransitData) (*core.Invalidation, bool, error)
	SUUID() core.SUUID
	SetSUUID(core.SUUID)
}

// BaseHeuristicOpt ... Functional option for BaseHeuristic
type BaseHeuristicOpt = func(bi *BaseHeuristic) *BaseHeuristic

// BaseHeuristic ... Base heuristic implementation
type BaseHeuristic struct {
	sUUID  core.SUUID
	inType core.RegisterType
}

// NewBaseHeuristic ... Initializer for BaseHeuristic
// This is a base type that's inherited by all hardcoded
// heuristic implementations
func NewBaseHeuristic(inType core.RegisterType,
	opts ...BaseHeuristicOpt) Heuristic {
	bi := &BaseHeuristic{
		inType: inType,
	}

	for _, opt := range opts {
		opt(bi)
	}

	return bi
}

// SUUID ... Returns the heuristic session UUID
func (bi *BaseHeuristic) SUUID() core.SUUID {
	return bi.sUUID
}

// InputType ... Returns the input type for the heuristic
func (bi *BaseHeuristic) InputType() core.RegisterType {
	return bi.inType
}

// Invalidate ... Invalidates the heuristic; defaults to no-op
func (bi *BaseHeuristic) Invalidate(core.TransitData) (*core.Invalidation, bool, error) {
	return nil, false, nil
}

// SetSUUID ... Sets the heuristic session UUID
func (bi *BaseHeuristic) SetSUUID(sUUID core.SUUID) {
	bi.sUUID = sUUID
}

// ValidateInput ... Validates the input type for the heuristic
func (bi *BaseHeuristic) ValidateInput(td core.TransitData) error {
	if td.Type != bi.InputType() {
		return fmt.Errorf(invalidInTypeErr, bi.InputType(), td.Type)
	}

	return nil
}
