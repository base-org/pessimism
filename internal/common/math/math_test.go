package math_test

import (
	"math/big"
	"testing"

	"github.com/base-org/pessimism/internal/common/math"
	"github.com/stretchr/testify/assert"
)

// Test_SorensonDice ... Tests Sorenson Dice similarity
func Test_SorensonDice(t *testing.T) {
	var tests = []struct {
		name     string
		function func(t *testing.T, a, b string, expected float64)
	}{
		{
			name: "Equal Strings",
			function: func(t *testing.T, a, b string, expected float64) {
				assert.Equal(t, math.SorensonDice(a, b), expected)
			},
		},
		{
			name: "Unequal Strings",
			function: func(t *testing.T, a, b string, expected float64) {
				assert.Equal(t, math.SorensonDice(a, b), expected)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.function(t, "0x123", "0x123", 1)
			test.function(t, "0x123", "0x124", 0.75)
		})
	}

}

func Test_PercentOf(t *testing.T) {
	assert := func(t *testing.T, a, b, expected *big.Float) {
		assert.Equal(t, math.PercentOf(a, b), expected)
	}

	assert(t, big.NewFloat(25), big.NewFloat(100), big.NewFloat(25))
	assert(t, big.NewFloat(10), big.NewFloat(100), big.NewFloat(10))
}
