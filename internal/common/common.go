package common

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
)

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
