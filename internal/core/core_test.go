package core_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_TransitData(t *testing.T) {
	// Verify construction
	td := core.NewTransitData(
		core.GethBlock,
		nil,
	)

	assert.NotNil(t, td, "TransitData should not be nil")
	assert.NotNil(t, td.Timestamp, "TransitData timestamp should not be nil")

	// Verify addressing
	addressed := td.Addressed()
	assert.False(t, addressed, "TransitData should not be addressed")

	td.Address = common.HexToAddress("0x456")
	addressed = td.Addressed()
	assert.True(t, addressed, "TransitData should be addressed")
}

func Test_EngineRelay(t *testing.T) {
	outChan := make(chan core.HeuristicInput)

	eir := core.NewEngineRelay(core.NilPUUID(), outChan)
	dummyTD := core.NewTransitData(core.AccountBalance, nil)

	// Verify relay and wrapping

	go func() {
		_ = eir.RelayTransitData(dummyTD)
	}()

	heurInput := <-outChan

	assert.NotNil(t, heurInput, "HeuristicInput should not be nil")
	assert.Equal(t, heurInput.PUUID, core.NilPUUID(), "HeuristicInput PUUID should be nil")
	assert.Equal(t, heurInput.Input, dummyTD, "HeuristicInput Input should be dummyTD")
}

func Test_SessionParams(t *testing.T) {
	isp := core.NewSessionParams()
	assert.NotNil(t, isp, "SessionParams should not be nil")

	isp.SetValue("tst", "tst")
	val, err := isp.Value("tst")
	assert.Nil(t, err, "Value should not return an error")
	assert.Equal(t, val, "tst", "Value should return the correct value")

	isp.SetNestedArg("bland(1,2,3)")
	nestedArgs := isp.NestedArgs()
	assert.Equal(t, nestedArgs, []interface{}{"bland(1,2,3)"}, "NestedArgs should return the correct value")

}

func Test_UnmarshalYaml(t *testing.T) {

	type test struct {
		TestKey  core.StringFromEnv `yaml:"test_key"`
		TestKey2 core.StringFromEnv `yaml:"test_key2"`
	}

	t.Setenv("test_key", "test_value")

	yml := []byte(`
test_key: ${test_key}
test_key2: "test_value2"
`)

	t1 := &test{}
	err := yaml.Unmarshal(yml, t1)
	assert.Nil(t, err, "Unmarshal should not return an error")
	assert.Equal(t, "test_value", t1.TestKey.String(), "Unmarshal should return the correct value")
	assert.Equal(t, "test_value2", t1.TestKey2.String(), "Unmarshal should return the correct value")
}
