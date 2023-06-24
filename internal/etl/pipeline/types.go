package pipeline

type ActivityState uint8

const (
	Booting ActivityState = iota
	Syncing
	Active
	Crashed
)

func (as ActivityState) String() string {
	switch as {
	case Booting:
		return "booting"

	case Syncing:
		return "syncing"

	case Active:
		return "active"

	case Crashed:
		return "crashed"
	}

	return "unknown"
}

const (
	// EtlStore error constants
	couldNotCastErr = "could not cast component initializer function to %s constructor type"
	pIDNotFoundErr  = "could not find pipeline ID for %s"
	uuidNotFoundErr = "could not find matching UUID for pipeline entry"

	// ComponentGraph error constants
	cUUIDNotFoundErr = "component with ID %s does not exist within component graph"
	cUUIDExistsErr   = "component with ID %s already exists in component graph"
	edgeExistsErr    = "edge already exists from (%s) to (%s) in component graph"

	// Manager error constants
	unknownCompType = "unknown component type %s provided"

	noAggregatorErr = "aggregator component has yet to be implemented"
)
