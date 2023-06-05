package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry/oracle"
	"github.com/base-org/pessimism/internal/etl/registry/pipe"
	"github.com/base-org/pessimism/internal/state"
)

const (
	noEntryErr = "could not find entry in registry for encoded register type %s"

	withAddressing = true
	noAddressing   = false

	withNesting = true
	noNesting   = false
)

// Registry ... Interface for registry
type Registry interface {
	GetDependencyPath(rt core.RegisterType) (core.RegisterDependencyPath, error)
	GetRegister(rt core.RegisterType) (*core.DataRegister, error)
}

// componentRegistry ...
type componentRegistry struct {
	registers map[core.RegisterType]*core.DataRegister
}

// NewRegistry ... Initializer
func NewRegistry() Registry {
	registers := map[core.RegisterType]*core.DataRegister{
		core.GethBlock: {
			Addressing:           noAddressing,
			DataType:             core.GethBlock,
			ComponentType:        core.Oracle,
			ComponentConstructor: oracle.NewGethBlockOracle,

			Dependencies: noDeps(),
			StateKeys:    noState(),
		},
		core.ContractCreateTX: {
			Addressing:           noAddressing,
			DataType:             core.ContractCreateTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: pipe.NewCreateContractTxPipe,

			Dependencies: makeDeps(core.GethBlock),
			StateKeys:    noState(),
		},
		core.BlackholeTX: {
			Addressing:           noAddressing,
			DataType:             core.BlackholeTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: pipe.NewBlackHoleTxPipe,

			Dependencies: makeDeps(core.GethBlock),
			StateKeys:    noState(),
		},
		core.AccountBalance: {
			Addressing:           withAddressing,
			DataType:             core.AccountBalance,
			ComponentType:        core.Oracle,
			ComponentConstructor: oracle.NewAddressBalanceOracle,

			// TODO() - Add dependency for geth block
			Dependencies: noDeps(),
			StateKeys: []core.StateKey{
				state.MakeKey(core.AddressPrefix, "addresses", noNesting),
			},
		},
		core.EventLog: {
			Addressing:           withAddressing,
			DataType:             core.EventLog,
			ComponentType:        core.Oracle,
			ComponentConstructor: oracle.NewEventOracle,

			// TODO() - Add dependency for geth block
			Dependencies: noDeps(),
			StateKeys: []core.StateKey{
				state.MakeKey(core.AddressPrefix, "addresses", withNesting),
				state.MakeKey(core.EventPrefix, "event_log", withNesting),
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

// noState ... Returns empty state keys, indicating no state dependencies
// for cross subsystem communication (i.e. ETL -> Risk Engine)
func noState() []core.StateKey {
	return []core.StateKey{}
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
