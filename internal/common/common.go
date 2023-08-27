package common

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

const (
	// Event declaration strings
	OutputProposedEvent   = "OutputProposed(bytes32,uint256,uint256,uint256)"
	WithdrawalProvenEvent = "WithdrawalProven(bytes32,address,address)"
	TransferEvent         = "Transfer(address,address,uint256)"
	ApprovalEvent         = "Approval(address,address,uint256)"
)

// MovementEvent ... Statically represents a transferal operation
// for some ERC20 token (i.e, transfer(), transferFrom(), approve())
type MovementEvent struct {
	Type   string
	From   common.Address
	To     common.Address
	Amount *big.Int
}

// ParseMovementEvent ... Parses a movement event from a slice of log topics
func ParseMovementEvent(topics []common.Hash) (*MovementEvent, error) {

	moveType := "unknown"
	if topics[0] == common.HexToHash(TransferEvent) {
		moveType = TransferEvent
	}

	if topics[0] == common.HexToHash(ApprovalEvent) {
		moveType = ApprovalEvent
	}

	if len(topics) != 4 {
		return nil, fmt.Errorf("invalid topics length, expected 4, got %d", len(topics))
	}
	return &MovementEvent{
		Type:   moveType,
		From:   common.HexToAddress(topics[1].Hex()),
		To:     common.HexToAddress(topics[2].Hex()),
		Amount: topics[3].Big(),
	}, nil
}

// WeiToEther ... Converts wei to ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

// SliceToAddresses ... Converts a slice of strings to a slice of addresses
func SliceToAddresses(slice []string) []common.Address {
	var addresses []common.Address
	for _, addr := range slice {
		addresses = append(addresses, common.HexToAddress(addr))
	}

	return addresses
}
