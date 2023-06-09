package registry_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/stretchr/testify/assert"
)

func Test_AddressPreprocess(t *testing.T) {
	isp := core.NewSessionParams()
	err := registry.AddressPreprocess(isp)
	assert.Error(t, err, "failure should occur when no address is provided")

	isp.SetValue(core.AddrKey, "0x69")

	err = registry.AddressPreprocess(isp)
	assert.NoError(t, err)
}

func Test_EventPreprocess(t *testing.T) {
	isp := core.NewSessionParams()
	err := registry.EventPreprocess(isp)
	assert.Error(t, err, "failure should occur when no address is provided")

	isp.SetValue(core.AddrKey, "0x69")
	err = registry.EventPreprocess(isp)
	assert.Error(t, err, "failure should occur when no event is provided")

	isp.SetNestedArg("transfer(address,address,uint256)")
	err = registry.EventPreprocess(isp)
	assert.Nil(t, err, "no error should occur when nested args are provided")
}

func Test_WithdrawEnforcePreprocess(t *testing.T) {
	isp := core.NewSessionParams()

	err := registry.WithdrawEnforcePreprocess(isp)
	assert.Error(t, err, "failure should occur when no l1_portal is provided")

	isp.SetValue(core.L1Portal, "0x69")
	err = registry.WithdrawEnforcePreprocess(isp)
	assert.Error(t, err, "failure should occur when no l2tol1 passer is provided")

	isp.SetValue(core.L2ToL1MessagePasser, "0x666")
	err = registry.WithdrawEnforcePreprocess(isp)
	assert.NoError(t, err)

	isp.SetNestedArg("transfer(address,address,uint256)")
	err = registry.WithdrawEnforcePreprocess(isp)
	assert.Error(t, err, "failure should when nested args are provided")

}

func Test_InvTable(t *testing.T) {
	tabl := registry.NewInvariantTable()

	for key, inv := range tabl {
		t.Run(key.String(), func(t *testing.T) {

			assert.NotNil(t, inv.Constructor)
			assert.NotNil(t, inv.Preprocess)
			assert.NotEqual(t, inv.InputType.String(), core.UnknownType)
		})
	}
}
