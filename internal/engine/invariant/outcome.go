package invariant

type OutcomeAction int

const (
	SlackPost OutcomeAction = iota + 1
)

type Outcome struct {
	Message string
	Action  OutcomeAction
}
