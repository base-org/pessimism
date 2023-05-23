package registry

import (
	"fmt"

	pess_core "github.com/base-org/pessimism/internal/core"
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

		Invariant: invariant.NewBaseInvariant(pess_core.ContractCreateTX),
	}
}

func (ei *ExampleInvariant) InputType() pess_core.RegisterType {
	return pess_core.ContractCreateTX
}

func (ei *ExampleInvariant) Invalidate(td pess_core.TransitData) (bool, error) {
	logging.NoContext().Debug("Checking invalidation")

	if td.Type != pess_core.ContractCreateTX {
		return false, fmt.Errorf("invalid type supplied")
	}

	tx, ok := td.Value.(*types.Transaction)
	if !ok {
		return false, fmt.Errorf("could not cast transit data to geth transaction type")
	}

	logging.NoContext().Info("Comparing addresses")
	from, err := types.Sender(types.LatestSignerForChainID(tx.ChainId()), tx)
	if err != nil {
		return false, err
	}

	logging.NoContext().Info("Comparing", zap.String("From", from.String()), zap.String("To", ei.cfg.FromAddress))
	result := (from == common.HexToAddress(ei.cfg.FromAddress))

	return result, nil
}
