package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

type Registry interface {
	GetRegister(rt core.RegisterType) (*core.DataRegister, error)
}

type componentRegistry struct {
	registers []*core.DataRegister
}

func NewRegistry() Registry {

	registers := []&core.DataRegister{

		&core.DataRegister{
			DataType:             core.GethBlock,
			ComponentType:        core.Oracle,
			ComponentConstructor: NewGethBlockOracle,
			Dependencies:         make([]*core.DataRegister, 0),
		},
		&core.DataRegister{
			DataType:             core.GethBlock,
			ComponentType:        core.Oracle,
			ComponentConstructor: NewGethBlockOracle,
			Dependencies:         make([]*core.DataRegister, 0),
		},
		&core.DataRegister{
			DataType:             core.ContractCreateTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: NewCreateContractTxPipe,
			Dependencies:         []*core.DataRegister{gethBlockReg},
		},
		&core.DataRegister{
			DataType:             core.BlackholeTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: NewBlackHoleTxPipe,
			Dependencies:         []*core.DataRegister{gethBlockReg},
		},
		&core.DataRegister{
			DataType:             core.ContractCreateTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: NewCreateContractTxPipe,
			Dependencies:         []*core.DataRegister{gethBlockReg},
		},
		&core.DataRegister{
			DataType:             core.BlackholeTX,
			ComponentType:        core.Pipe,
			ComponentConstructor: NewBlackHoleTxPipe,
			Dependencies:         []*core.DataRegister{gethBlockReg},
		}
}

	return &componentRegistry{registers}
}

func (cr *componentRegistry) GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	switch rt {
	case core.GethBlock:
		return gethBlockReg, nil

	case core.ContractCreateTX:
		return contractCreateTXReg, nil

	case core.BlackholeTX:
		return blackHoleTxReg, nil

	default:
		return nil, fmt.Errorf("no register could be found for type: %s", rt)
	}
}
