package registry

import (
	"fmt"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"go.uber.org/zap"
)

type ExampleInvConfig struct {
	FromAddress string `json:"from"`
}

type ExampleInvariant struct {
	cfg *ExampleInvConfig

	invariant.Invariant
}

func NewExampleInvariant(cfg *ExampleInvConfig) invariant.Invariant {
	return &ExampleInvariant{
		cfg: cfg,

		Invariant: invariant.NewBaseInvariant(core.ContractCreateTX),
	}
}

func (ei *ExampleInvariant) InputType() core.RegisterType {
	return core.ContractCreateTX
}

func (ei *ExampleInvariant) Invalidate(td core.TransitData) (*core.InvalOutcome, bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != core.ContractCreateTX {
		return nil, false, fmt.Errorf("invalid type supplied")
	}

	tx, ok := td.Value.(*types.Transaction)
	if !ok {
		return nil, false, fmt.Errorf("could not cast transit data to geth transaction type")
	}

	logging.NoContext().Info("Comparing addresses")
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return nil, false, err
	}

	logging.NoContext().Info("Comparing", zap.String("From", from.String()), zap.String("To", ei.cfg.FromAddress))
	if from == common.HexToAddress(ei.cfg.FromAddress) {
		return &core.InvalOutcome{
			TimeStamp: time.Now(),
			Message:   fmt.Sprintf("Creation tx detected from %s", from.String()),
		}, true, nil
	}

	return nil, false, nil
}
