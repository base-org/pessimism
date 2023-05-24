package registry

import (
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
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

	invariant.Invariant
}

func NewInvocationTrackerInvariant(cfg *InvocationInvConfig) invariant.Invariant {
	return &InvocationTrackerInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(core.GethBlock),
	}
}

func (it *InvocationTrackerInvariant) InputType() core.RegisterType {
	return core.GethBlock
}

func (it *InvocationTrackerInvariant) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != core.GethBlock {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	block, ok := td.Value.(types.Block)
	if !ok {
		return nil, false, fmt.Errorf("could not cast transit data to geth Block type")
	}

	for _, tx := range block.Transactions() {
		from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
		if err != nil {
			return nil, false, err
		}

		if from == common.HexToAddress(it.cfg.FromAddress) {
			return &core.InvalOutcome{
				TimeStamp: time.Now(),
				Message:   fmt.Sprintf("Invocation detected from %s", from.String()),
				SUUID:     it.UUID(),
			}, true, nil
		}
	}
	return nil, false, nil
}
