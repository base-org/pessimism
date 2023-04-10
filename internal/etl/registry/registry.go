package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
)

var (
	gethBlockReg = &core.DataRegister{
		DataType:             core.GethBlock,
		ComponentType:        core.Oracle,
		ComponentConstructor: NewGethBlockOracle,
		Dependencies:         make([]*core.DataRegister, 0),
	}

	contractCreateTXReg = &core.DataRegister{
		DataType:             core.ContractCreateTX,
		ComponentType:        core.Pipe,
		ComponentConstructor: NewCreateContractTxPipe,
		Dependencies:         []*core.DataRegister{gethBlockReg},
	}

	blackHoleTxReg = &core.DataRegister{
		DataType:             core.BlackholeTX,
		ComponentType:        core.Pipe,
		ComponentConstructor: NewBlackHoleTxPipe,
		Dependencies:         []*core.DataRegister{gethBlockReg},
	}
)

func GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
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
