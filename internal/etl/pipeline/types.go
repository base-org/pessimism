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
