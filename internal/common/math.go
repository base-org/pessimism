package common

import "math/big"

// PercentOf - calculate what percent [number1] is of [number2].
// ex. 300 is 12.5% of 2400
func PercentOf(part, total *big.Float) *big.Float {
	x0 := new(big.Float).Mul(part, big.NewFloat(100))
	return new(big.Float).Quo(x0, total)
}
