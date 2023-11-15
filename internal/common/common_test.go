package common_test

import (
	"math/big"
	"testing"

	"github.com/base-org/pessimism/internal/common"
	"github.com/base-org/pessimism/internal/common/math"
	geth_common "github.com/ethereum/go-ethereum/common"

	"github.com/stretchr/testify/assert"
)

const (
	weiPerETH = 1000000000000000000
)

// Test_WeiToEth ... Tests wei to ether conversion
func Test_WeiToEth(t *testing.T) {
	ether := math.WeiToEther(big.NewInt(weiPerETH))
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
