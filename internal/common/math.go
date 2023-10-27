package common

import "math/big"

// PercentOf - calculate what percent x0 is of x1.
func PercentOf(part, total *big.Float) *big.Float {
	whole := 100.0
	x0 := new(big.Float).Mul(part, big.NewFloat(whole))
	return new(big.Float).Quo(x0, total)
}

// SorensonDice - calculate the Sorenson-Dice coefficient between two strings.
func SorensonDice(s1, s2 string) float64 {
	// https://en.wikipedia.org/wiki/S%C3%B8rensen%E2%80%93Dice_coefficient
	// 2 * |X & Y| / (|X| + |Y|)

	var (
		n1 = len(s1)
		n2 = len(s2)
	)

	if n1 == 0 || n2 == 0 {
		return 0
	}

	intersects := 0
	for i := 0; i < n1-1; i++ {
		a := s1[i : i+2]
		for j := 0; j < n2-1; j++ {
			b := s2[j : j+2]
			if a == b {
				intersects++
				break
			}
		}
	}

	return float64(2*intersects) / float64(n1+n2-2)
}
