package common

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

// WeiToEther ... Converts wei to ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}
