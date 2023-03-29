package registry

import (
	"fmt"

	"github.com/base-org/pessimism/internal/conduit/models"
)

const (
	GETH_BLOCK models.RegisterType = "GETH_BLOCK"
)

var (
	geth_block_register = &DataRegister{
		DataType:             GETH_BLOCK,
		ComponentType:        models.Oracle,
		ComponentConstructor: NewGethBlockOracle,
		Dependencies:         make([]DataRegister, 0),
	}
)

type DataRegister struct {
	DataType             models.RegisterType
	ComponentType        models.ComponentType
	ComponentConstructor interface{}
	// TODO - Introduce dependency management logic
	Dependencies []DataRegister
}

func GetRegister(rt models.RegisterType) (*DataRegister, error) {

	switch rt {
	case GETH_BLOCK:
		return geth_block_register, nil

	default:
		return nil, fmt.Errorf("No register could be found for type: %s", rt)
	}

}
