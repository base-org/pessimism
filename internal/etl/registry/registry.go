package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/models"
)

const (
	GethBlock        models.RegisterType = "GETH_BLOCK"
	ContractCreateTX models.RegisterType = "CONTRACT_CREATE_TX"
	BlackholeTX      models.RegisterType = "BLACK_HOLE_TX"
)

var (
	gethBlockReg = &DataRegister{
		DataType:             GethBlock,
		ComponentType:        models.Oracle,
		ComponentConstructor: NewGethBlockOracle,
		Dependencies:         make([]*DataRegister, 0),
	}

	contractCreateTXReg = &DataRegister{
		DataType:             ContractCreateTX,
		ComponentType:        models.Pipe,
		ComponentConstructor: NewCreateContractTxPipe,
		Dependencies:         []*DataRegister{gethBlockReg},
	}

	blackHoleTxReg = &DataRegister{
		DataType:             BlackholeTX,
		ComponentType:        models.Pipe,
		ComponentConstructor: NewBlackHoleTxPipe,
		Dependencies:         []*DataRegister{gethBlockReg},
	}
)

type DataRegister struct {
	DataType             models.RegisterType
	ComponentType        models.ComponentType
	ComponentConstructor interface{}
	// TODO - Introduce dependency management logic
	Dependencies []*DataRegister
}

func GetRegister(rt models.RegisterType) (*DataRegister, error) {
	switch rt {
	case GethBlock:
		return gethBlockReg, nil

	case ContractCreateTX:
		return contractCreateTXReg, nil

	default:
		return nil, fmt.Errorf("no register could be found for type: %s", rt)
	}
}
