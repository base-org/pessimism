package oracle

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	pess_common "github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"

	"go.uber.org/zap"
)

// TODO(#21): Verify config validity during Oracle construction
// AddressBalanceODef ... Address register oracle definition used to drive oracle component
type AddressBalanceODef struct {
	pUUID      core.PUUID
	cfg        *core.ClientConfig
	client     client.EthClientInterface
	currHeight *big.Int
	sk         *core.StateKey
}

// NewAddressBalanceODef ... Initializer for address.balance oracle definition
func NewAddressBalanceODef(cfg *core.ClientConfig, client client.EthClientInterface,
	h *big.Int) *AddressBalanceODef {
	return &AddressBalanceODef{
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

// NewAddressBalanceOracle ... Initializer for address.balance oracle component
func NewAddressBalanceOracle(ctx context.Context, cfg *core.ClientConfig,
	opts ...component.Option) (component.Component, error) {
	client, err := client.FromContext(ctx, cfg.Network)
	if err != nil {
		return nil, err
	}

	od := NewAddressBalanceODef(cfg, client, nil)
	o, err := component.NewOracle(ctx, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	od.sk = o.StateKey().Clone()
	return o, nil
}

// ConfigureRoutine ... Sets up the oracle client connection and persists puuid to definition state
func (oracle *AddressBalanceODef) ConfigureRoutine(pUUID core.PUUID) error {
	oracle.pUUID = pUUID
	return nil
}

// BackTestRoutine ...
// NOTE - This oracle does not support backtesting
// TODO (#59) : Add account balance backtesting support
func (oracle *AddressBalanceODef) BackTestRoutine(_ context.Context, _ chan core.TransitData,
	_ *big.Int, _ *big.Int) error {
	return fmt.Errorf(noBackTestSupportError)
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client for address (EOA, Contract) native balance amounts
func (oracle *AddressBalanceODef) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	stateStore, err := state.FromContext(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond) //nolint:durationcheck // inapplicable
	for {
		select {
		case <-ticker.C: // Polling
			logging.NoContext().Debug("Getting addresess",
				zap.String(core.PUUIDKey, oracle.pUUID.String()))

			// Get addresses from shared state store for pipeline uuid

			addresses, err := stateStore.GetSlice(ctx, oracle.sk)
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			for _, address := range addresses {
				// Convert to go-ethereum address type
				gethAddress := common.HexToAddress(address)
				logging.NoContext().Debug("Balance query",
					zap.String(core.AddrKey, gethAddress.String()))

				// Get balance using go-ethereum client
				weiBalance, err := oracle.client.BalanceAt(ctx, gethAddress, nil)
				if err != nil {
					logging.WithContext(ctx).Error(err.Error())
					continue
				}

				// Convert wei to ether
				// NOTE - There is a possibility of precision loss here
				// TODO (#58) : Verify precision loss
				ethBalance, _ := pess_common.WeiToEther(weiBalance).Float64()

				logging.NoContext().Debug("Balance",
					zap.String(core.AddrKey, gethAddress.String()),
					zap.Int64("wei balance ", weiBalance.Int64()))

				logging.NoContext().Debug("Balance",
					zap.String(core.AddrKey, gethAddress.String()),
					zap.Float64("balance", ethBalance))

				// Send parsed float64 balance value to downstream component channel
				componentChan <- core.NewTransitData(core.AccountBalance, ethBalance,
					core.WithAddress(gethAddress))
			}

		case <-ctx.Done(): // Shutdown
			return nil
		}
	}
}
