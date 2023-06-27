package e2e_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/e2e"

	"github.com/base-org/pessimism/internal/api/models"
)

// Test_Balance_Enforcement
func Test_Balance_Enforcement(t *testing.T) {

	ts := e2e.CreateL2TestSuite(t)
	defer ts.Close()

	alice := ts.L2Cfg.Secrets.Addresses().Alice
	bob := ts.L2Cfg.Secrets.Addresses().Bob

	// Deploy a balance enforcement invariant session for Alice.
	err := ts.App.BootStrap([]models.InvRequestParams{{
		Network:      "layer2",
		PType:        "live",
		InvType:      "balance_enforcement",
		StartHeight:  nil,
		EndHeight:    nil,
		AlertingDest: "slack",
		SessionParams: map[string]interface{}{
			"address": alice.String(),
			"lower":   3, // i.e. alert if balance is less than 3 ETH
		},
	}})

	assert.NoError(t, err, "Failed to bootstrap balance enforcement invariant session")

	// Get Alice's balance.
	aliceAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), alice, nil)
	assert.NoError(t, err, "Failed to get Alice's balance")

	// Determine the gas cost of the transaction.
	gasAmt := 1_000_001
	bigAmt := big.NewInt(1_000_001)
	gasPrice := big.NewInt(int64(ts.L2Cfg.DeployConfig.L2GenesisBlockGasLimit))

	gasCost := gasPrice.Mul(gasPrice, bigAmt)

	signer := types.LatestSigner(ts.L2Geth.L2ChainConfig)
	drainAliceTx := types.MustSignNewTx(ts.L2Cfg.Secrets.Alice, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(ts.L2Cfg.DeployConfig.L2ChainID)),
		Nonce:     0,
		GasTipCap: big.NewInt(100),
		GasFeeCap: big.NewInt(100000),
		Gas:       uint64(gasAmt),
		To:        &bob,
		// Subtract the gas cost from the amount sent to Bob.
		Value: aliceAmt.Sub(aliceAmt, gasCost),
		Data:  nil,
	})

	assert.Equal(t, len(ts.Slack.SlackAlerts), 0, "No alerts should be sent before the transaction is sent")

	// Send the transaction to drain Alice's account of almost all ETH.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainAliceTx)
	assert.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change and send a notification to the mocked Slack server.
	time.Sleep(1 * time.Second)

	// Check that the balance enforcement was triggered using the mocked server cache.
	posts := ts.Slack.SlackAlerts

	assert.Greater(t, len(posts), 0, "No balance enforcement alert was sent")
	assert.Contains(t, posts[0].Text, "balance_enforcement", "Balance enforcement alert was not sent")

	// Get Bobs's balance.
	bobAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), bob, nil)
	assert.NoError(t, err, "Failed to get Alice's balance")

	// Create a transaction to send the ETH back to Alice.
	drainBobTx := types.MustSignNewTx(ts.L2Cfg.Secrets.Bob, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(ts.L2Cfg.DeployConfig.L2ChainID)),
		Nonce:     0,
		GasTipCap: big.NewInt(100),
		GasFeeCap: big.NewInt(100000),
		Gas:       uint64(gasAmt),
		To:        &alice,
		Value:     bobAmt.Sub(bobAmt, gasCost),
		Data:      nil,
	})

	// Send the transaction to redispurse the ETH from Bob back to Alice.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainBobTx)
	assert.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change.
	time.Sleep(1 * time.Second)

	// Empty the mocked Slack server cache.
	ts.Slack.ClearAlerts()

	// Wait to ensure that no new alerts are sent.
	time.Sleep(1 * time.Second)

	// Ensure that no new alerts were sent.
	assert.Equal(t, len(ts.Slack.SlackAlerts), 0, "No alerts should be sent after the transaction is sent")
}

func Test_Contract_Event(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t)
	defer ts.Close()

	l1Client := ts.Sys.Clients["l1"]

	// The string declaration of the event we want to listen for.
	updateSig := "ConfigUpdate(uint256,uint8,bytes)"

	// Deploy a contract event invariant session for the L1 system config addresss.
	err := ts.App.BootStrap([]models.InvRequestParams{{
		Network:      "layer1",
		PType:        "live",
		InvType:      "contract_event",
		StartHeight:  nil,
		EndHeight:    nil,
		AlertingDest: "slack",
		SessionParams: map[string]interface{}{
			"address": predeploys.DevSystemConfigAddr.String(),
			"args":    []interface{}{updateSig},
		},
	}})
	assert.NoError(t, err, "Error bootstrapping invariant session")

	// Get bindings for the L1 system config contract.
	sysconfig, err := bindings.NewSystemConfig(predeploys.DevSystemConfigAddr, l1Client)
	assert.NoError(t, err, "Error getting system config")

	// Obtain our signer.
	opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.SysCfgOwner, ts.Cfg.L1ChainIDBig())
	assert.NoError(t, err, "Error getting system config owner pk")

	// Assign arbitrary gas config values.
	overhead := big.NewInt(10000)
	scalar := big.NewInt(1)

	// Call setGasConfig method on the L1 system config contract.
	tx, err := sysconfig.SetGasConfig(opts, overhead, scalar)
	assert.NoError(t, err, "Error setting gas config")

	// Wait for the transaction to be canonicalized.
	txTimeoutDuration := 10 * time.Duration(ts.Cfg.DeployConfig.L1BlockTime) * time.Second
	receipt, err := e2e.WaitForTransaction(tx.Hash(), l1Client, txTimeoutDuration)

	assert.NoError(t, err, "Error waiting for transaction")
	assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	// Wait for Pessimism to process the newly emitted event and send a notification to the mocked Slack server.
	time.Sleep(1 * time.Second)
	posts := ts.Slack.SlackAlerts

	assert.Equal(t, len(posts), 1, "No system contract event alert was sent")
	assert.Contains(t, posts[0].Text, "contract_event", "System contract event alert was not sent")
}
