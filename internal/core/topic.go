package core

type TopicType uint8

const (
	BlockHeader TopicType = iota + 1
	Log
)

func (rt TopicType) String() string {
	switch rt {
	case BlockHeader:
		return "block_header"

	case Log:
		return "log"
	}

	return UnknownType
}

type DataTopic struct {
	Addressing bool
	Sk         *StateKey

	DataType     TopicType
	ProcessType  ProcessType
	Constructor  interface{}
	Dependencies []TopicType
}

func (dt *DataTopic) StateKey() *StateKey {
	return dt.Sk.Clone()
}

func (dt *DataTopic) Stateful() bool {
	return dt.Sk != nil
}

// Represents an inclusive acyclic sequential path of data register dependencies
type TopicPath struct {
	Path []*DataTopic
}

func (tp TopicPath) GeneratePathID(pt PathType, n Network) PathID {
	proc1, proc2 := tp.Path[0], tp.Path[len(tp.Path)-1]
	id1 := MakeProcessID(pt, proc1.ProcessType, proc1.DataType, n)
	id2 := MakeProcessID(pt, proc2.ProcessType, proc2.DataType, n)

	return MakePathID(pt, id1, id2)
}
