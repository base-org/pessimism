package core

// RegisterType ... One byte register type enum
type RegisterType uint8

const (
	AccountBalance RegisterType = iota + 1
	GethBlock
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

	case EventLog:
		return "event_log"
	}

	return UnknownType
}

// DataRegister ... Represents an ETL subsytem data type that
// can be produced and consumed by heterogenous components
type DataRegister struct {
	Addressing bool
	Sk         *StateKey

	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	Dependencies         []RegisterType
}

// StateKey ... Returns a cloned state key for a data register
func (dr *DataRegister) StateKey() *StateKey {
	return dr.Sk.Clone()
}

// Stateful ... Indicates whether the data register has statefulness
func (dr *DataRegister) Stateful() bool {
	return dr.Sk != nil
}

// RegisterDependencyPath ... Represents an inclusive acyclic sequential
// path of data register dependencies
type RegisterDependencyPath struct {
	Path []*DataRegister
}

// GeneratePUUID ... Generates a PUUID for an existing dependency path
// provided an enumerated pipeline and network type
func (rdp RegisterDependencyPath) GeneratePUUID(pt PipelineType, n Network) PUUID {
	firstComp, lastComp := rdp.Path[0], rdp.Path[len(rdp.Path)-1]
	firstUUID := MakeCUUID(pt, firstComp.ComponentType, firstComp.DataType, n)
	lastUUID := MakeCUUID(pt, lastComp.ComponentType, lastComp.DataType, n)

	return MakePUUID(pt, firstUUID, lastUUID)
}
