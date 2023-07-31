package invariant

import "github.com/base-org/pessimism/internal/core"

// DeployConfig ... Configuration for deploying an invariant session
type DeployConfig struct {
	Stateful bool
	StateKey *core.StateKey

	Network core.Network
	PUUID   core.PUUID
	Reuse   bool

	InvType   core.HeuristicType
	InvParams *core.SessionParams

	AlertDest core.AlertDestination
}
