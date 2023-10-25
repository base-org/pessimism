package e2e_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/base-org/pessimism/e2e"
	"github.com/base-org/pessimism/internal/api/models"
	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
)

// TestMultiDirectiveRouting ... Tests the E2E flow of a contract event heuristic with high priority alerts all
// necessary destinations
func TestMultiDirectiveRouting(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t)

	updateSig := "ConfigUpdate(uint256,uint8,bytes)"
	alertMsg := "System config gas config updated"

	_, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer1.String(),
		PType:         core.Live.String(),
		HeuristicType: core.ContractEvent.String(),
		StartHeight:   nil,
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			Msg: alertMsg,
			Sev: core.HIGH.String(),
		},
		SessionParams: map[string]interface{}{
			"address": ts.Cfg.L1Deployments.SystemConfigProxy.String(),
			"args":    []interface{}{updateSig},
		},
	}})

	assert.NoError(t, err, "Error bootstrapping heuristic session")

	sysCfg, err := bindings.NewSystemConfig(ts.Cfg.L1Deployments.SystemConfigProxy, ts.L1Client)
	assert.NoError(t, err, "Error getting system config")

	opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.SysCfgOwner, ts.Cfg.L1ChainIDBig())
	assert.NoError(t, err, "Error getting system config owner pk")

	overhead := big.NewInt(10000)
	scalar := big.NewInt(1)

	tx, err := sysCfg.SetGasConfig(opts, overhead, scalar)
	assert.NoError(t, err, "Error setting gas config")

	txTimeoutDuration := 10 * time.Duration(ts.Cfg.DeployConfig.L1BlockTime) * time.Second
	receipt, err := e2e.WaitForTransaction(tx.Hash(), ts.L1Client, txTimeoutDuration)

	assert.NoError(t, err, "Error waiting for transaction")
	assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	time.Sleep(3 * time.Second)
	slackPosts := ts.TestSlackSvr.SlackAlerts()
	pdPosts := ts.TestPagerDutyServer.PagerDutyAlerts()

	// Expect 2 alerts to each destination as alert-routing-cfg.yaml has two slack and two pagerduty destinations
	assert.Equal(t, 2, len(slackPosts), "Incorrect Number of slack posts sent")
	assert.Equal(t, 2, len(pdPosts), "Incorrect Number of pagerduty posts sent")
	assert.Contains(t, slackPosts[0].Text, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, slackPosts[1].Text, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, pdPosts[0].Payload.Summary, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, pdPosts[1].Payload.Summary, "contract_event", "System contract event alert was not sent")
}

// TestCoolDown ... Tests the E2E flow of a single
// balance enforcement heuristic session on L2 network with a cooldown.
func TestCoolDown(t *testing.T) {

	ts := e2e.CreateL2TestSuite(t)
	defer ts.Close()

	alice := ts.L2Cfg.Secrets.Addresses().Alice
	bob := ts.L2Cfg.Secrets.Addresses().Bob

	alertMsg := "one baby to another says:"
	// Deploy a balance enforcement heuristic session for Alice using a cooldown.
	_, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer2.String(),
		PType:         core.Live.String(),
		HeuristicType: core.BalanceEnforcement.String(),
		StartHeight:   nil,
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			// Set a cooldown to one minute.
			CoolDown: 60,
			Sev:      core.LOW.String(),
			Msg:      alertMsg,
		},
		SessionParams: map[string]interface{}{
			"address": alice.String(),
			"lower":   3, // i.e. alert if balance is less than 3 ETH
		},
	}})

	assert.NoError(t, err, "Failed to bootstrap balance enforcement heuristic session")

	// Get Alice's balance.
	aliceAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), alice, nil)
	assert.NoError(t, err, "Failed to get Alice's balance")

	// Determine the gas cost of the transaction.
	gasAmt := 1_000_001
	bigAmt := big.NewInt(1_000_001)
	gasPrice := big.NewInt(int64(ts.L2Cfg.DeployConfig.L2GenesisBlockGasLimit))

	gasCost := gasPrice.Mul(gasPrice, bigAmt)

	signer := types.LatestSigner(ts.L2Geth.L2ChainConfig)

	// Create a transaction from Alice to Bob that will drain almost all of Alice's ETH.
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

	// Send the transaction to drain Alice's account of almost all ETH.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainAliceTx)
	assert.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change and send a notification to the mocked Slack server.
	time.Sleep(2 * time.Second)

	// Check that the balance enforcement was triggered using the mocked server cache.
	posts := ts.TestSlackSvr.SlackAlerts()

	assert.Equal(t, 1, len(posts), "No balance enforcement alert was sent")
	assert.Contains(t, posts[0].Text, "balance_enforcement", "Balance enforcement alert was not sent")
	assert.Contains(t, posts[0].Text, alertMsg)

	// Ensure that no new alerts are sent for provided cooldown period.
	time.Sleep(1 * time.Second)
	posts = ts.TestSlackSvr.SlackAlerts()
	assert.Equal(t, 1, len(posts), "No alerts should be sent after the transaction is sent")
}
