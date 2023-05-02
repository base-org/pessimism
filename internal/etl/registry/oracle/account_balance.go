package oracle

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"
)

// TODO(#21): Verify config validity during Oracle construction
// AddressBalanceODef ... Address register oracle definition used to drive oracle component
type AddressBalanceODef struct {
	id         core.ComponentUUID
	cfg        *config.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewAddressBalanceODef ... Initializer for address.balance oracle definition
func NewAddressBalanceODef(cfg *config.OracleConfig, client client.EthClientInterface, h *big.Int) *AddressBalanceODef {
	return &AddressBalanceODef{
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

// NewAddressBalanceOracle ... Initializer for address.balance oracle component
func NewAddressBalanceOracle(ctx context.Context, ot core.PipelineType,
	cfg *config.OracleConfig, opts ...component.Option) (component.Component, error) {
	client := client.NewEthClient()

	od := NewAddressBalanceODef(cfg, client, nil)
	o, err := component.NewOracle(ctx, ot, core.GethBlock, od, opts...)
	if err != nil {
		return nil, err
	}

	od.id = o.ID()

	return o, nil

}

func (oracle *AddressBalanceODef) ConfigureRoutine() error {
	ctxTimeout, ctxCancel := context.WithTimeout(context.Background(),
		time.Second*time.Duration(core.EthClientTimeout))
	defer ctxCancel()

	logging.WithContext(ctxTimeout).Info("Setting up GETH Block client")

	err := oracle.client.DialContext(ctxTimeout, oracle.cfg.RPCEndpoint)

	if err != nil {
		return err
	}
	return nil
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

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			addresses, err := stateStore.Get(ctx, oracle.id.String())
			if err != nil {
				return err
			}

			for _, address := range addresses {
				gethAddress := common.HexToAddress(address)
				balance, err := oracle.client.BalanceAt(ctx, gethAddress, nil)
				if err != nil {
					logging.WithContext(ctx).Error(err.Error())
				}

				componentChan <- core.TransitData{
					Timestamp: time.Now(),
					Type:      core.AccountBalance,
					Value: core.AccountBalanceVal{
						Address: gethAddress,
						Balance: balance,
					},
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}
