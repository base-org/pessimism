package invariant

import "github.com/base-org/pessimism/internal/core"

// DeployConfig ... Configuration for deploying an invariant session
type DeployConfig struct {
	Network   core.Network
	PUUID     core.PUUID
	InvType   core.InvariantType
	InvParams core.InvSessionParams
	Register  *core.DataRegister
}
