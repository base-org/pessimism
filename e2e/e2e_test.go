package e2e_test

import (
	"context"
	"crypto/ecdsa"
	"math/big"

	"github.com/ethereum/go-ethereum/rpc"

	"testing"
	"time"

	"github.com/ethereum-optimism/optimism/op-bindings/bindings"
	"github.com/ethereum-optimism/optimism/op-bindings/predeploys"

	"github.com/ethereum-optimism/optimism/op-node/withdrawals"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/e2e"

	"github.com/base-org/pessimism/internal/api/models"
)

// Test_Balance_Enforcement ... Tests the E2E flow of a single
// balance enforcement invariant session on L2 network.
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

	assert.Equal(t, len(ts.TestSvr.SlackAlerts()), 0, "No alerts should be sent before the transaction is sent")

	// Send the transaction to drain Alice's account of almost all ETH.
	_, err = ts.L2Geth.AddL2Block(context.Background(), drainAliceTx)
	assert.NoError(t, err, "Failed to create L2 block with transaction")

	// Wait for Pessimism to process the balance change and send a notification to the mocked Slack server.
	time.Sleep(1 * time.Second)

	// Check that the balance enforcement was triggered using the mocked server cache.
	posts := ts.TestSvr.SlackAlerts()

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
	ts.TestSvr.ClearAlerts()

	// Wait to ensure that no new alerts are sent.
	time.Sleep(1 * time.Second)

	// Ensure that no new alerts were sent.
	assert.Equal(t, len(ts.TestSvr.Payloads), 0, "No alerts should be sent after the transaction is sent")
}

// Test_Contract_Event ... Tests the E2E flow of a single
// contract event invariant session on L1 network.
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
	posts := ts.TestSvr.SlackAlerts()

	assert.Equal(t, len(posts), 1, "No system contract event alert was sent")
	assert.Contains(t, posts[0].Text, "contract_event", "System contract event alert was not sent")
}

// TestAccount defines an account generated b
type TestAccount struct {
	Key    *ecdsa.PrivateKey
	Addr   common.Address
	L1Opts *bind.TransactOpts
	L2Opts *bind.TransactOpts
}

// Test_Withdrawal_Enforcement ...
func Test_Withdrawal_Enforcement(t *testing.T) {

	ts := e2e.CreateSysTestSuite(t)
	ts.Cfg.DeployConfig.FinalizationPeriodSeconds = 6
	defer ts.Close()

	// Obtain our sequencer, verifier, and transactor keypair.
	l1Client := ts.Sys.Clients["l1"]
	l2Seq := ts.Sys.Clients["sequencer"]
	l2Verif := ts.Sys.Clients["verifier"]

	// Define our L1 transaction timeout duration.
	txTimeoutDuration := 10 * time.Duration(ts.Cfg.DeployConfig.L1BlockTime) * time.Second

	// Bind to the deposit contract.
	depositContract, err := bindings.NewOptimismPortal(predeploys.DevOptimismPortalAddr, l1Client)
	_ = depositContract
	assert.NoError(t, err)

	// Create a test account state for our transactor.
	transactorKey := ts.Cfg.Secrets.Alice
	transactor := TestAccount{
		Key:    transactorKey,
		L1Opts: nil,
		L2Opts: nil,
	}

	transactor.L1Opts, err = bind.NewKeyedTransactorWithChainID(transactor.Key, ts.Cfg.L1ChainIDBig())
	assert.NoError(t, err)
	transactor.L2Opts, err = bind.NewKeyedTransactorWithChainID(transactor.Key, ts.Cfg.L2ChainIDBig())
	assert.NoError(t, err)

	// Bind to the L2-L1 message passer.
	l2l1MessagePasser, err := bindings.NewL2ToL1MessagePasser(predeploys.L2ToL1MessagePasserAddr, l2Seq)
	assert.NoError(t, err, "error binding to message passer on L2")

	// Deploy a dummy L2ToL1 message passer for testing.
	fakeAddr, tx, _, err := bindings.DeployL2ToL1MessagePasser(transactor.L2Opts, l2Seq)
	assert.NoError(t, err, "error deploying message passer on L2")

	_, err = e2e.WaitForTransaction(tx.Hash(), l2Seq, txTimeoutDuration)
	assert.NoError(t, err, "error waiting for transaction")

	// Determine the address our request will come from.
	fromAddr := crypto.PubkeyToAddress(transactor.Key.PublicKey)

	// Initiate Withdrawal.
	withdrawAmount := big.NewInt(500_000_000_000)
	transactor.L2Opts.Value = withdrawAmount
	tx, err = l2l1MessagePasser.InitiateWithdrawal(transactor.L2Opts, fromAddr, big.NewInt(21000), nil)
	assert.Nil(t, err, "sending initiate withdraw tx")

	// Wait for the transaction to appear in the L2 verifier.
	receipt, err := e2e.WaitForTransaction(tx.Hash(), l2Verif, txTimeoutDuration)
	assert.Nil(t, err, "withdrawal initiated on L2 sequencer")
	assert.Equal(t, receipt.Status, types.ReceiptStatusSuccessful, "transaction failed")

	// Wait for the finalization period, then we can finalize this withdrawal.
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Duration(ts.Cfg.DeployConfig.L1BlockTime)*time.Second)
	blockNumber, err := withdrawals.WaitForFinalizationPeriod(ctx, l1Client, predeploys.DevOptimismPortalAddr, receipt.BlockNumber)
	cancel()
	assert.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), txTimeoutDuration)
	header, err := l2Verif.HeaderByNumber(ctx, new(big.Int).SetUint64(blockNumber))
	cancel()
	assert.Nil(t, err)

	l2OutputOracle, err := bindings.NewL2OutputOracleCaller(predeploys.DevL2OutputOracleAddr, l1Client)
	assert.Nil(t, err)

	rpcClient, err := rpc.Dial(ts.Sys.Nodes["verifier"].WSEndpoint())

	assert.Nil(t, err)
	proofCl := gethclient.New(rpcClient)
	receiptCl := ethclient.NewClient(rpcClient)

	// Now create the withdrawal
	params, err := withdrawals.ProveWithdrawalParameters(context.Background(), proofCl, receiptCl, tx.Hash(), header, l2OutputOracle)
	assert.Nil(t, err)

	// Obtain our withdrawal parameters
	withdrawalTransaction := &bindings.TypesWithdrawalTransaction{
		Nonce:    params.Nonce,
		Sender:   params.Sender,
		Target:   params.Target,
		Value:    params.Value,
		GasLimit: params.GasLimit,
		Data:     params.Data,
	}
	l2OutputIndexParam := params.L2OutputIndex
	outputRootProofParam := params.OutputRootProof
	withdrawalProofParam := params.WithdrawalProof

	// Setup Pessimism to listen for fraudulent withdrawals
	// We use two invariants here; one configured with a dummy L1 message passer
	// and one configured with the real L1->L2 message passer contract. This allows us to
	// ensure that an alert is only produced using faulty message passer.
	err = ts.App.BootStrap([]models.InvRequestParams{{
		// This is the one that should produce an alert
		Network:      "layer1",
		PType:        "live",
		InvType:      "withdrawal_enforcement",
		StartHeight:  nil,
		EndHeight:    nil,
		AlertingDest: "slack",
		SessionParams: map[string]interface{}{
			"l1_portal":   predeploys.DevOptimismPortal,
			"l2_messager": fakeAddr.String(),
		},
	},
		{
			// This is the one that shouldn't produce an alert
			Network:      "layer1",
			PType:        "live",
			InvType:      "withdrawal_enforcement",
			StartHeight:  nil,
			EndHeight:    nil,
			AlertingDest: "slack",
			SessionParams: map[string]interface{}{
				"l1_portal":   predeploys.DevOptimismPortal,
				"l2_messager": predeploys.L2ToL1MessagePasserAddr.String(),
			},
		},
	})
	assert.NoError(t, err, "Error bootstrapping invariant session")
	time.Sleep(1 * time.Second)

	// Prove withdrawal. This checks the proof so we only finalize if this succeeds
	tx, err = depositContract.ProveWithdrawalTransaction(
		transactor.L1Opts,
		*withdrawalTransaction,
		l2OutputIndexParam,
		outputRootProofParam,
		withdrawalProofParam,
	)
	assert.Nil(t, err, "withdrawal should successfully prove")

	// Wait for the transaction to appear in L1
	_, err = e2e.WaitForTransaction(tx.Hash(), l1Client, txTimeoutDuration)
	assert.Nil(t, err, "withdrawal finalized on L1")
	time.Sleep(1 * time.Second)

	// Ensure Pessimism has detected what it considers a "faulty" withdrawal
	alerts := ts.TestSvr.SlackAlerts()
	assert.Equal(t, 1, len(alerts), "expected 1 alert")
	assert.Contains(t, alerts[0].Text, "withdrawal_enforcement", "expected alert to be for withdrawal_enforcement")
	assert.Contains(t, alerts[0].Text, fakeAddr.String(), "expected alert to be for dummy L2ToL1MessagePasser")
}
