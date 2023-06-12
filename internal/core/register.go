package core

// RegisterType ... One byte register type enum
type RegisterType uint8

const (
	AccountBalance RegisterType = iota + 1
	GethBlock
	ContractCreateTX
	BlackholeTX
)

// String ... Returns string representation of a
// register enum
func (rt RegisterType) String() string {
	switch rt {
	case AccountBalance:
		return "account.balance"

	case GethBlock:
		return "geth.block"

	case ContractCreateTX:
		return "contract.create.tx"

	case BlackholeTX:
		return "blackhole.tx"
	}

	return UnknownType
}

// DataRegister ... Represents an ETL subsytem data type that
// can be produced and consumed by heterogenous components
type DataRegister struct {
	Addressing bool

	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	Dependencies         []RegisterType
}

// RegisterDependencyPath ... Represents an inclusive acyclic sequential
// path of data register dependencies
type RegisterDependencyPath struct {
	Path []*DataRegister
}

// GeneratePipelineUUID ... Generates a pipelineUUID for an existing dependency path
// provided an enumerated pipeline and network type
func (rdp RegisterDependencyPath) GeneratePipelineUUID(pt PipelineType, n Network) PipelineUUID {
	firstComp, lastComp := rdp.Path[0], rdp.Path[len(rdp.Path)-1]
	firstUUID := MakeComponentUUID(pt, firstComp.ComponentType, firstComp.DataType, n)
	lastUUID := MakeComponentUUID(pt, lastComp.ComponentType, lastComp.DataType, n)

	return MakePipelineUUID(pt, firstUUID, lastUUID)
}
