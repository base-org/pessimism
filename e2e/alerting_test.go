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
	"github.com/ethereum-optimism/optimism/op-e2e/e2eutils/wait"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// These are localstack specific Topic ARNs that are used to test the SNS integration.
	MultiDirectiveTopicArn = "arn:aws:sns:us-east-1:000000000000:multi-directive-test-topic"
	CoolDownTopicArn       = "arn:aws:sns:us-east-1:000000000000:alert-cooldown-test-topic"
)

// TestMultiDirectiveRouting ... Tests the E2E flow of a contract event heuristic with high priority alerts all
// necessary destinations
func TestMultiDirectiveRouting(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t, MultiDirectiveTopicArn)
	defer ts.Close()

	updateSig := "ConfigUpdate(uint256,uint8,bytes)"
	alertMsg := "System config gas config updated"

	ids, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer1.String(),
		HeuristicType: core.ContractEvent.String(),
		StartHeight:   nil,
		EndHeight:     nil,
		AlertingParams: &core.AlertPolicy{
			Msg: alertMsg,
			Sev: core.HIGH.String(), // The use of HIGH priority should trigger all alert destinations
		},
		SessionParams: map[string]interface{}{
			"address": ts.Cfg.L1Deployments.SystemConfigProxy.String(),
			"args":    []interface{}{updateSig},
		},
	}})

	require.Len(t, ids, 1, "Incorrect number of heuristic sessions created")
	require.NoError(t, err, "Error bootstrapping heuristic session")

	sysCfg, err := bindings.NewSystemConfig(ts.Cfg.L1Deployments.SystemConfigProxy, ts.L1Client)
	require.NoError(t, err, "Error getting system config")

	opts, err := bind.NewKeyedTransactorWithChainID(ts.Cfg.Secrets.SysCfgOwner, ts.Cfg.L1ChainIDBig())
	require.NoError(t, err, "Error getting system config owner pk")

	overhead := big.NewInt(10000)
	scalar := big.NewInt(1)

	tx, err := sysCfg.SetGasConfig(opts, overhead, scalar)
	require.NoError(t, err, "Error setting gas config")

	receipt, err := wait.ForReceipt(context.Background(), ts.L1Client, tx.Hash(), types.ReceiptStatusSuccessful)

	require.NoError(t, err, "Error waiting for transaction")
	require.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	// Wait for Pessimism to process the newly emitted event and send a notification to the mocked Slack
	// and PagerDuty servers.
	require.NoError(t, wait.For(context.Background(), 500*time.Millisecond, func() (bool, error) {
		id := ids[0].PathID
		height, err := ts.Subsystems.PathHeight(id)
		if err != nil {
			return false, err
		}

		return height != nil && height.Uint64() > receipt.BlockNumber.Uint64(), nil
	}))

	snsMessages, err := e2e.GetSNSMessages(ts.AppCfg.AlertConfig.SNSConfig.Endpoint, "multi-directive-test-queue")
	require.NoError(t, err)

	assert.Len(t, snsMessages.Messages, 1, "Incorrect number of SNS messages sent")
	assert.Contains(t, *snsMessages.Messages[0].Body, "contract_event", "System contract event alert was not sent")

	slackPosts := ts.TestSlackSvr.SlackAlerts()
	pdPosts := ts.TestPagerDutyServer.PagerDutyAlerts()

	// Expect 2 alerts to each destination as alert-routing-cfg.yaml has two slack and two pagerduty destinations
	require.Equal(t, 2, len(slackPosts), "Incorrect Number of slack posts sent")
	require.Equal(t, 2, len(pdPosts), "Incorrect Number of pagerduty posts sent")

	assert.Contains(t, slackPosts[0].Text, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, slackPosts[1].Text, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, pdPosts[0].Payload.Summary, "contract_event", "System contract event alert was not sent")
	assert.Contains(t, pdPosts[1].Payload.Summary, "contract_event", "System contract event alert was not sent")
}

// TestCoolDown ... Tests the E2E flow of a single
// balance enforcement heuristic session on L2 network with a cooldown.
func TestCoolDown(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t, CoolDownTopicArn)
	defer ts.Close()

	alice := ts.Cfg.Secrets.Addresses().Alice
	bob := ts.Cfg.Secrets.Addresses().Bob

	alertMsg := "one baby to another says:"
	// Deploy a balance enforcement heuristic session for Alice using a cooldown.
	ids, err := ts.App.BootStrap([]*models.SessionRequestParams{{
		Network:       core.Layer2.String(),
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

	require.NoError(t, err, "Failed to bootstrap balance enforcement heuristic session")

	// Get Alice's balance.
	aliceAmt, err := ts.L2Client.BalanceAt(context.Background(), alice, nil)
	require.NoError(t, err, "Failed to get Alice's balance")

	// Determine the gas cost of the transaction.
	gasAmt := 1_000_001
	bigAmt := big.NewInt(1_000_001)
	gasPrice := big.NewInt(int64(ts.Cfg.DeployConfig.L2GenesisBlockGasLimit))

	gasCost := gasPrice.Mul(gasPrice, bigAmt)

	signer := types.LatestSigner(ts.Sys.L2GenesisCfg.Config)

	// Create a transaction from Alice to Bob that will drain almost all of Alice's ETH.
	drainAliceTx := types.MustSignNewTx(ts.Cfg.Secrets.Alice, signer, &types.DynamicFeeTx{
		ChainID:   big.NewInt(int64(ts.Cfg.DeployConfig.L2ChainID)),
		Nonce:     0,
		GasTipCap: big.NewInt(100),
		GasFeeCap: big.NewInt(100000),
		Gas:       uint64(gasAmt),
		To:        &bob,
		// Subtract the gas cost from the amount sent to Bob.
		Value: aliceAmt.Sub(aliceAmt, gasCost),
		Data:  nil,
	})

	err = ts.L2Client.SendTransaction(context.Background(), drainAliceTx)
	require.NoError(t, err)

	receipt, err := wait.ForReceipt(context.Background(), ts.L2Client, drainAliceTx.Hash(), types.ReceiptStatusSuccessful)
	require.NoError(t, err)

	require.NoError(t, wait.For(context.Background(), 1000*time.Millisecond, func() (bool, error) {
		id := ids[0].PathID
		height, err := ts.Subsystems.PathHeight(id)
		if err != nil {
			return false, err
		}

		return height != nil && height.Uint64() > receipt.BlockNumber.Uint64(), nil
	}))

	// Check that the balance enforcement was triggered using the mocked server cache.
	posts := ts.TestSlackSvr.SlackAlerts()

	sqsMessages, err := e2e.GetSNSMessages(ts.AppCfg.AlertConfig.SNSConfig.Endpoint, "alert-cooldown-test-queue")
	assert.NoError(t, err)
	assert.Len(t, sqsMessages.Messages, 1, "Incorrect number of SNS messages sent")
	assert.Contains(t, *sqsMessages.Messages[0].Body, "balance_enforcement", "Balance enforcement alert was not sent")

	require.Equal(t, 1, len(posts), "No balance enforcement alert was sent")
	assert.Contains(t, posts[0].Text, "balance_enforcement", "Balance enforcement alert was not sent")
	assert.Contains(t, posts[0].Text, alertMsg)

	// Ensure that no new alerts are sent for provided cooldown period.
	time.Sleep(1 * time.Second)
	posts = ts.TestSlackSvr.SlackAlerts()
	assert.Equal(t, 1, len(posts), "No alerts should be sent after the transaction is sent")

}
