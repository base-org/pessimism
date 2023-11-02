package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

const (
	noEntryErr = "could not find entry in registry for encoded register type %s"
)

// Registry ... Interface for registry
type Registry interface {
	GetDependencyPath(rt core.RegisterType) (core.RegisterDependencyPath, error)
	GetRegister(rt core.RegisterType) (*core.DataRegister, error)
}

// componentRegistry ... Registry implementation
type componentRegistry struct {
	registers map[core.RegisterType]*core.DataRegister
}

// NewRegistry ... Instantiates a new hardcoded registry
// that contains all extractable ETL data types
func NewRegistry() Registry {
	registers := map[core.RegisterType]*core.DataRegister{
		core.BlockHeader: {
			Addressing:    false,
			DataType:      core.BlockHeader,
			ComponentType: core.Reader,
			Constructor:   NewHeaderTraversal,

			Dependencies: noDeps(),
			Sk:           noState(),
		},

		core.EventLog: {
			Addressing:    true,
			DataType:      core.EventLog,
			ComponentType: core.Transformer,
			Constructor:   NewEventParserPipe,

			Dependencies: makeDeps(core.BlockHeader),
			Sk: &core.StateKey{
				Nesting: true,
				Prefix:  core.EventLog,
				ID:      core.AddressKey,
				PUUID:   nil,
			},
		},
	}

	return &componentRegistry{registers}
}

// makeDeps ... Makes dependency slice
func makeDeps(types ...core.RegisterType) []core.RegisterType {
	deps := make([]core.RegisterType, len(types))
	copy(deps, types)

	return deps
}

// noDeps ... Returns empty dependency slice
func noDeps() []core.RegisterType {
	return []core.RegisterType{}
}

// noState ... Returns empty state key, indicating no state dependencies
// for cross subsystem communication (i.e. ETL -> Risk Engine)
func noState() *core.StateKey {
	return nil
}

// GetDependencyPath ... Returns in-order slice of ETL pipeline path
func (cr *componentRegistry) GetDependencyPath(rt core.RegisterType) (core.RegisterDependencyPath, error) {
	destRegister, err := cr.GetRegister(rt)
	if err != nil {
		return core.RegisterDependencyPath{}, err
	}

	registers := make([]*core.DataRegister, len(destRegister.Dependencies)+1)

	registers[0] = destRegister

	for i, depType := range destRegister.Dependencies {
		depRegister, err := cr.GetRegister(depType)
		if err != nil {
			return core.RegisterDependencyPath{}, err
		}

		registers[i+1] = depRegister
	}

	return core.RegisterDependencyPath{Path: registers}, nil
}

// GetRegister ... Returns a data register provided an enum type
func (cr *componentRegistry) GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	if _, exists := cr.registers[rt]; !exists {
		return nil, fmt.Errorf(noEntryErr, rt)
	}

	return cr.registers[rt], nil
}
