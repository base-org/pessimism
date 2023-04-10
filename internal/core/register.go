package core

type RegisterType uint8

const (
	GethBlock RegisterType = iota + 1
	ContractCreateTX
	BlackholeTX
)

func (rt RegisterType) String() string {
	switch rt {
	case GethBlock:
		return "GETH_BLOCK"

	case ContractCreateTX:
		return "CONTRACT_CREATE_TX"

	case BlackholeTX:
		return "BLACKHOLE_TX"
	}

	return "unknown"
}

type DataRegister struct {
	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	// TODO - Introduce dependency management logic
	Dependencies []*DataRegister
}
