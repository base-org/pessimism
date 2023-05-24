package oracle

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"go.uber.org/zap"
)

// TODO(#21): Verify config validity during Oracle construction
// AddressBalanceODef ... Address register oracle definition used to drive oracle component
type AddressBalanceODef struct {
	pUUID      core.PipelineUUID
	cfg        *core.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewAddressBalanceODef ... Initializer for address.balance oracle definition
func NewAddressBalanceODef(cfg *core.OracleConfig, client client.EthClientInterface, h *big.Int) *AddressBalanceODef {
	return &AddressBalanceODef{
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

// NewAddressBalanceOracle ... Initializer for address.balance oracle component
func NewAddressBalanceOracle(ctx context.Context, ot core.PipelineType,
	cfg *core.OracleConfig, opts ...component.Option) (component.Component, error) {
	client := client.NewEthClient()

	od := NewAddressBalanceODef(cfg, client, nil)
	o, err := component.NewOracle(ctx, ot, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	return o, nil
}

func weiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// ConfigureRoutine ... Sets up the oracle client connection and persists puuid to defintion state
func (oracle *AddressBalanceODef) ConfigureRoutine(cUUID core.ComponentUUID, pUUID core.PipelineUUID) error {
	oracle.pUUID = pUUID

	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(core.EthClientTimeout))
	defer ctxCancel()

	logging.WithContext(ctxTimeout).Info("Setting up Account Balance client")

	return oracle.client.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)
}

// BackTestRoutine ...
func (oracle *AddressBalanceODef) BackTestRoutine(ctx context.Context, componentChan chan core.TransitData,
	startHeight *big.Int, endHeight *big.Int) error {
	return fmt.Errorf(noBackTestSupportError)
}

// ReadRoutine ... Sequentially polls go-ethereum compatible execution
// client for address (EOA, Contract) native balance amounts
func (oracle *AddressBalanceODef) ReadRoutine(ctx context.Context, componentChan chan core.TransitData) error {
	stateStore, err := state.FromContext(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(oracle.cfg.PollInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			logging.NoContext().Debug("Getting addresess",
				zap.String(core.PUUIDKey, oracle.pUUID.String()))
			addresses, err := stateStore.Get(ctx, oracle.pUUID.String())
			if err != nil {
				logging.WithContext(ctx).Error(err.Error())
				continue
			}

			for _, address := range addresses {
				gethAddress := common.HexToAddress(address)

				logging.NoContext().Debug("Balance query",
					zap.String(core.AddrKey, gethAddress.String()))
				weiBalance, err := oracle.client.BalanceAt(ctx, gethAddress, nil)
				if err != nil {
					logging.WithContext(ctx).Error(err.Error())
					continue
				}

				// Convert wei to ether
				ethBalance, _ := weiToEther(weiBalance).Float64()

				logging.NoContext().Debug("Balance",
					zap.String(core.AddrKey, gethAddress.String()),
					zap.Int64("wei balance ", weiBalance.Int64()))

				logging.NoContext().Debug("Balance",
					zap.String(core.AddrKey, gethAddress.String()),
					zap.Float64("balance", ethBalance))

				componentChan <- core.NewTransitData(core.AccountBalance, ethBalance,
					core.WithAddress(gethAddress))
			}

		case <-ctx.Done():
			return nil
		}
	}
}
