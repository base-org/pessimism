package core

// RegisterType ... One byte register type enum
type RegisterType uint8

const (
	GethBlock RegisterType = iota + 1
	ContractCreateTX
	BlackholeTX
)

// String ... Returns string representation of a
// register enum
func (rt RegisterType) String() string {
	switch rt {
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
	DataType             RegisterType
	ComponentType        ComponentType
	ComponentConstructor interface{}
	Dependencies         []*DataRegister
}

type RegisterDependencyPath struct {
	regPath []*DataRegister
}

func (rdp RegisterDependencyPath) OutputRegister() *DataRegister {
	return rdp.regPath[len(rdp.regPath)-1]
}

func (rdp RegisterDependencyPath) GeneratePipelineUUID(pt PipelineType, n Network) PipelineUUID {
	firstComp, lastComp := rdp.regPath[0], rdp.OutputRegister()
	firstUUID := MakeComponentUUID(pt, firstComp.ComponentType, firstComp.DataType, n)
	lastUUID := MakeComponentUUID(pt, lastComp.ComponentType, lastComp.DataType, n)

	return MakePipelineUUID(pt, firstUUID, lastUUID)
}

// GetDependencyPath ... Returns inclusive dependency path for data register
func (dr *DataRegister) GetDependencyPath() RegisterDependencyPath {

	registers := append([]*DataRegister{dr}, dr.Dependencies...)

	return RegisterDependencyPath{registers}
}
