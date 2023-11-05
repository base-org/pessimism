package etl

type ActivityState uint8

const (
	INACTIVE ActivityState = iota
	ACTIVE
	CRASHED
	TERMINATED
)

func (as ActivityState) String() string {
	switch as {
	case INACTIVE:
		return "inactive"

	case ACTIVE:
		return "active"

	case CRASHED:
		return "crashed"

	case TERMINATED:
		return "terminated"
	}

	return "unknown"
}

const (
	// EtlStore error constants
	couldNotCastErr = "could not cast process initializer function to %s constructor type"
	pIDNotFoundErr  = "could not find path ID for %s"
	uuidNotFoundErr = "could not find matching UUID for path entry"

	// ProcessGraph error constants
	cUUIDNotFoundErr = "process with ID %s does not exist within dag"
	procExistsErr    = "process with ID %s already exists in dag"
	edgeExistsErr    = "edge already exists from (%s) to (%s) in dag"

	emptyPathError = "path must contain at least one process"
	// Manager error constants
	unknownCompType = "unknown process type %s provided"
)
