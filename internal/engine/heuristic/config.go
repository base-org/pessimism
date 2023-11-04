package heuristic

import "github.com/base-org/pessimism/internal/core"

// DeployConfig ... Configuration for deploying a heuristic session
type DeployConfig struct {
	Stateful bool
	StateKey *core.StateKey

	Network core.Network
	PathID  core.PathID
	Reuse   bool

	HeuristicType core.HeuristicType
	Params        *core.SessionParams

	AlertingPolicy *core.AlertPolicy
}
