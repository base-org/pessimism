package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry/oracle"
	"github.com/base-org/pessimism/internal/etl/registry/pipe"
)

var (
	gethBlockReg = &core.DataRegister{
		DataType:             core.GethBlock,
		ComponentType:        core.Oracle,
		ComponentConstructor: oracle.NewGethBlockOracle,
		Dependencies:         make([]*core.DataRegister, 0),
	}

	contractCreateTXReg = &core.DataRegister{
		DataType:             core.ContractCreateTX,
		ComponentType:        core.Pipe,
		ComponentConstructor: pipe.NewCreateContractTxPipe,
		Dependencies:         []*core.DataRegister{gethBlockReg},
	}

	blackHoleTxReg = &core.DataRegister{
		DataType:             core.BlackholeTX,
		ComponentType:        core.Pipe,
		ComponentConstructor: pipe.NewBlackHoleTxPipe,
		Dependencies:         []*core.DataRegister{gethBlockReg},
	}

	accountBalanceReg = &core.DataRegister{
		DataType:             core.AccountBalance,
		ComponentType:        core.Oracle,
		ComponentConstructor: oracle.NewAddressBalanceOracle,
		Dependencies:         []*core.DataRegister{},
	}
)

// GetRegister ... Returns a register entry value provided a valid register type
func GetRegister(rt core.RegisterType) (*core.DataRegister, error) {
	switch rt {
	case core.GethBlock:
		return gethBlockReg, nil

	case core.ContractCreateTX:
		return contractCreateTXReg, nil

	case core.BlackholeTX:
		return blackHoleTxReg, nil

	case core.AccountBalance:
		return accountBalanceReg, nil

	default:
		return nil, fmt.Errorf("no register could be found for type: %s", rt)
	}
}
