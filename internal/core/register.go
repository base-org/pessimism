package core

// RegisterType ... One byte register type enum
type RegisterType uint8

const (
	AccountBalance RegisterType = iota + 1
	GethBlock
	ContractCreateTX
	BlackholeTX
	EventLog
)

// String ... Returns string representation of a
// register enum
func (rt RegisterType) String() string {
	switch rt {
	case AccountBalance:
		return "account_balance"

	case GethBlock:
		return "geth_block"

	// TODO - Deprecate
	case ContractCreateTX:
		return "contract.create.tx"

	// TODO - Deprecate
	case BlackholeTX:
		return "blackhole.tx"

	case EventLog:
		return "event_log"
	}

	return UnknownType
}

// DataRegister ... Represents an ETL subsytem data type that
// can be produced and consumed by heterogenous components
type DataRegister struct {
	Addressing bool
	StateKeys  []StateKey

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
