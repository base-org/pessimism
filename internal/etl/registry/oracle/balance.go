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
	"github.com/ethereum/go-ethereum/common"
)

// TODO(#21): Verify config validity during Oracle construction
// AddressBalanceODef ...GethBlock register oracle definition used to drive oracle component
type AddressBalanceODef struct {
	addresses  map[common.Address]interface{}
	cfg        *config.OracleConfig
	client     client.EthClientInterface
	currHeight *big.Int
}

// NewAddressBalanceODef ... Initializer for address.balance oracle definition
func NewAddressBalanceODef(cfg *config.OracleConfig, client client.EthClientInterface, h *big.Int) *AddressBalanceODef {
	return &AddressBalanceODef{
		addresses:  make(map[common.Address]interface{}, 0),
		cfg:        cfg,
		client:     client,
		currHeight: h,
	}
}

func (oracle *AddressBalanceODef) addAddress(addr common.Address) error {
	if _, exists := oracle.addresses[addr]; exists {
		return fmt.Errorf("Address balance is already being monitored")
	}

	oracle.addresses[addr] = nil
	return nil
}

func (oracle *AddressBalanceODef) removeAddress(addr common.Address) error {
	if _, exists := oracle.addresses[addr]; !exists {
		return fmt.Errorf("Address is not being monitored")
	}

	delete(oracle.addresses, addr)
	return nil
}

// func (oracle *AddressBalanceODef) HandleUpdate(any) error {
// 	if _, exists := oracle.addresses[addr]; !exists {
// 		return fmt.Errorf("Address is not being monitored")
// 	}

// 	delete(oracle.addresses, addr)
// 	return nil
// }

// NewAddressBalanceOracle ... Initializer for address.balance oracle component
func NewAddressBalanceOracle(ctx context.Context, ot core.PipelineType,
	cfg *config.OracleConfig, opts ...component.Option) (component.Component, error) {
	client := client.NewEthClient()
	od := NewAddressBalanceODef(cfg, client, nil)

	return component.NewOracle(ctx, ot, core.GethBlock, od, opts...)
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

	ticker := time.NewTicker(pollInterval * time.Millisecond)
	for {
		select {
		case <-ticker.C:

			for address := range oracle.addresses {
				balance, err := oracle.client.BalanceAt(ctx, address, nil)
				if err != nil {
					logging.WithContext(ctx).Error(err.Error())
				}

				componentChan <- core.TransitData{
					Timestamp: time.Now(),
					Type:      core.AccountBalance,
					Value: core.AccountBalanceVal{
						Address: address,
						Balance: balance,
					},
				}
			}

		case <-ctx.Done():
			return nil
		}
	}
}
