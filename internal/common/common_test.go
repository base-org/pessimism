package common

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	weiPerETH = 1000000000000000000
)

// Test_WeiToEth ... Tests wei to ether conversion
func Test_WeiToEth(t *testing.T) {
	ether := WeiToEther(big.NewInt(weiPerETH))
	etherFloat, _ := ether.Float64()

	assert.Equal(t, etherFloat, float64(1), "should be equal")
}
