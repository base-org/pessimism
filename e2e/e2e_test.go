package e2e_test

import (
	"context"
	"math/big"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/e2e"

	"github.com/base-org/pessimism/internal/api/models"
)

// Test_Balance_Enforcement
func Test_Balance_Enforcement(t *testing.T) {

	ts := e2e.CreateTestSuite(t)

	alice := ts.L2Cfg.Secrets.Addresses().Alice
	bob := ts.L2Cfg.Secrets.Addresses().Bob

	// Deploy a balance enforcement invariant session for Alice
	err := ts.App.BootStrap([]models.InvRequestParams{{
		Network:      "layer2",
		PType:        "live",
		InvType:      "balance_enforcement",
		StartHeight:  nil,
		EndHeight:    nil,
		AlertingDest: "slack",
		SessionParams: map[string]interface{}{
			"address": alice.String(),
			"lower":   3, // Alert if balance is less than 3 ETH
		},
	}})

	if err != nil {
		t.Fatal(err)
	}

	// Get Alice's balance
	aliceAmt, err := ts.L2Geth.L2Client.BalanceAt(context.Background(), alice, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Determine the gas cost of the transaction
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
		// Subtract the gas cost from the amount sent to Bob
		Value: aliceAmt.Sub(aliceAmt, gasCost),
		Data:  nil,
	})

	assert.Equal(t, len(ts.Slack.SlackAlerts), 0, "No alerts should be sent before the transaction is sent")

	// Send the transaction to drain Alice's account of almost all ETH
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainAliceTx)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for Pessimism to process the balance change and send a notification to the mocked Slack server
	time.Sleep(1 * time.Second)

	// Check that the balance enforcement was triggered
	posts := ts.Slack.SlackAlerts

	assert.Greater(t, len(posts), 0, "No balance enforcement alert was sent")
	assert.Contains(t, posts[0].Text, "balance_enforcement", "Balance enforcement alert was not sent")
}
