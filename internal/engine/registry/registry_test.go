package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/stretchr/testify/assert"
)

func Test_AddressPreprocess(t *testing.T) {
	isp := core.NewSessionParams(core.Layer2)
	err := registry.ValidateAddressing(isp)
	assert.Error(t, err, "failure should occur when no address is provided")

	isp.SetValue(logging.AddrKey, "0x69")

	err = registry.ValidateAddressing(isp)
	assert.NoError(t, err)
}

func Test_EventPreprocess(t *testing.T) {
	isp := core.NewSessionParams(core.Layer1)
	err := registry.ValidateTracking(isp)
	assert.Error(t, err, "failure should occur when no address is provided")

	isp.SetValue(logging.AddrKey, "0x69")
	err = registry.ValidateTracking(isp)
	assert.Error(t, err, "failure should occur when no event is provided")

	isp.SetNestedArg("transfer(address,address,uint256)")
	err = registry.ValidateTracking(isp)
	assert.Nil(t, err, "no error should occur when nested args are provided")
}

func TestUnsafeWithdrawPrepare(t *testing.T) {
	isp := core.NewSessionParams(core.Layer1)

	err := registry.WithdrawHeuristicPrep(isp)
	assert.Error(t, err, "failure should occur when no l1_portal is provided")

	isp.SetValue(core.L1Portal, "0x69")
	err = registry.WithdrawHeuristicPrep(isp)
	assert.Error(t, err, "failure should occur when no l2tol1 passer is provided")

	isp.SetValue(core.L2ToL1MessagePasser, "0x666")
	err = registry.WithdrawHeuristicPrep(isp)
	assert.NoError(t, err)

	isp.SetNestedArg("transfer(address,address,uint256)")
	err = registry.WithdrawHeuristicPrep(isp)
	assert.Error(t, err, "failure should when nested args are provided")

}

func Test_InvTable(t *testing.T) {
	tabl := registry.NewHeuristicTable()

	for key, h := range tabl {
		t.Run(key.String(), func(t *testing.T) {

			assert.NotNil(t, h.Constructor)
			assert.NotNil(t, h.PrepareValidate)
			assert.NotEqual(t, h.InputType.String(), core.UnknownType)
		})
	}
}
