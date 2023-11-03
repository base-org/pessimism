package registry_test

import (
	"context"
	"math/big"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_Balance_Assess(t *testing.T) {
	upper := float64(5)
	lower := float64(1)

	ctx, ms := mocks.Context(context.Background(), gomock.NewController(t))

	bi, err := registry.NewBalanceHeuristic(ctx,
		&registry.BalanceInvConfig{
			Address:    "0x123",
			UpperBound: &upper,
			LowerBound: &lower,
		})

	assert.NoError(t, err)

	num := big.NewInt(1)
	// No activation
	testData1 := core.TransitData{
		Network: core.Layer1,
		Type:    core.BlockHeader,
		Value: types.Header{
			Number: num,
		},
	}

	ms.MockL1.EXPECT().
		BalanceAt(ctx, common.HexToAddress("0x123"), num).Return(big.NewInt(3000000000000000000), nil).Times(1)
	as, err := bi.Assess(testData1)
	assert.NoError(t, err)
	assert.False(t, as.Activated())

	// Upper bound activation
	num = num.Add(num, big.NewInt(1))
	testData2 := core.TransitData{
		Network: core.Layer1,
		Type:    core.BlockHeader,
		Value: types.Header{
			Number: num,
		},
	}

	ms.MockL1.EXPECT().
		BalanceAt(ctx, common.HexToAddress("0x123"), num).Return(big.NewInt(6000000000000000000), nil).Times(1)

	as, err = bi.Assess(testData2)
	assert.NoError(t, err)
	assert.True(t, as.Activated())

	num = num.Add(num, big.NewInt(1))
	// Lower bound activation
	testData3 := core.TransitData{
		Network: core.Layer1,
		Type:    core.BlockHeader,
		Value: types.Header{
			Number: num,
		},
	}
	ms.MockL1.EXPECT().
		BalanceAt(ctx, common.HexToAddress("0x123"), num).Return(big.NewInt(600000000000000000), nil).Times(1)

	as, err = bi.Assess(testData3)
	assert.NoError(t, err)
	assert.True(t, as.Activated())
}
