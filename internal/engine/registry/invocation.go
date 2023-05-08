package registry

import (
	"fmt"

	pess_core "github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type InvocationInvConfig struct {
	FromAddress string `json:"from"`
}

type InvocationTrackerInvariant struct {
	cfg *InvocationInvConfig
}

func NewInvocationTrackerInvariant(cfg *InvocationInvConfig) invariant.Invariant {
	return &InvocationTrackerInvariant{
		cfg: cfg,
	}
}

func (it *InvocationTrackerInvariant) InputType() pess_core.RegisterType {
	return pess_core.GethBlock
}

func (it *InvocationTrackerInvariant) Invalidate(td pess_core.TransitData) (bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != pess_core.GethBlock {
		return false, fmt.Errorf("invalid type supplied")
	}

	block, ok := td.Value.(types.Block)
	if !ok {
		return false, fmt.Errorf("could not cast transit data to geth Block type")
	}

	for _, tx := range block.Transactions() {
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			return false, err
		}

		if from == common.HexToAddress(it.cfg.FromAddress) {
			return true, nil
		}
	}
	return false, nil
}
