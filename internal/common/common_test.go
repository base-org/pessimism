package common_test

import (
	"math/big"
	"testing"

	"github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/core"
	geth_common "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
)

const (
	weiPerETH = 1000000000000000000
)

// Test_WeiToEth ... Tests wei to ether conversion
func Test_WeiToEth(t *testing.T) {
	ether := common.WeiToEther(big.NewInt(weiPerETH))
	etherFloat, _ := ether.Float64()

	assert.Equal(t, etherFloat, float64(1), "should be equal")
}

func Test_SliceToAddresses(t *testing.T) {
	addresses := make([]string, 0)
	addresses = append(addresses, "0x00000000")
	addresses = append(addresses, "0x00000001")

	convertedAddresses := common.SliceToAddresses(addresses)
	assert.Equal(t, convertedAddresses,
		[]geth_common.Address{geth_common.HexToAddress("0x00000000"), geth_common.HexToAddress("0x00000001")})

}

// Test_DLQ ... Tests all DLQ functionality
func Test_DLQ(t *testing.T) {
	dlq := common.NewTransitDLQ(5)

	// A. Add 5 elements and test size
	for i := 0; i < 5; i++ {
		td := core.NewTransitData(core.RegisterType(0), nil)

		err := dlq.Add(&td)
		assert.NoError(t, err)
	}

	// B. Add 6th element and test error
	td := core.NewTransitData(core.RegisterType(0), nil)
	err := dlq.Add(&td)

	assert.Error(t, err)

	// C. Pop 1 element and test size
	elem, err := dlq.Pop()
	assert.Equal(t, elem.Type, core.RegisterType(0))
	assert.NoError(t, err)

	// D. Pop all elements and test size
	entries := dlq.PopAll()
	assert.Equal(t, len(entries), 4)
	assert.True(t, dlq.Empty(), true)
}

// Test_SorensonDice ... Tests Sorenson Dice similarity
func Test_SorensonDice(t *testing.T) {
	var tests = []struct {
		name     string
		function func(t *testing.T, a, b string, expected float64)
	}{
		{
			name: "Equal Strings",
			function: func(t *testing.T, a, b string, expected float64) {
				assert.Equal(t, common.SorensonDice(a, b), expected)
			},
		},
		{
			name: "Unequal Strings",
			function: func(t *testing.T, a, b string, expected float64) {
				assert.Equal(t, common.SorensonDice(a, b), expected)
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
