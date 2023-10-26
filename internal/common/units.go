package common

import (
	"math/big"

	"github.com/ethereum/go-ethereum/params"
)

func WeiToGwei(i *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(i), big.NewFloat(params.GWei))
}

// WeiToEther ... Converts wei to ether
func WeiToEther(wei *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(wei), big.NewFloat(params.Ether))
}

func GweiToWei(i *big.Int) *big.Float {
	return new(big.Float).Mul(new(big.Float).SetInt(i), big.NewFloat(params.GWei))
}

func GweiToEth(i *big.Int) *big.Float {
	return new(big.Float).Quo(new(big.Float).SetInt(i), big.NewFloat(params.GWei))
}

func EtherToWei(i *big.Int) *big.Float {
	return new(big.Float).Mul(new(big.Float).SetInt(i), big.NewFloat(params.Ether))
}

func EtherToGwei(i *big.Int) *big.Float {
	return new(big.Float).Mul(new(big.Float).SetInt(i), big.NewFloat(params.GWei))
}
