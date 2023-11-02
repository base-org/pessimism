package common

import (
	"github.com/ethereum/go-ethereum/common"
)

// SliceToAddresses ... Converts a slice of strings to a slice of addresses
func SliceToAddresses(slice []string) []common.Address {
	var addresses []common.Address
	for _, addr := range slice {
		addresses = append(addresses, common.HexToAddress(addr))
	}

	return addresses
}
