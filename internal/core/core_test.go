package core_test

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/base-org/pessimism/internal/core"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/assert"
)

func Test_Event(t *testing.T) {
	// Verify construction
	e := core.NewEvent(
		core.BlockHeader,
		nil,
	)

	assert.NotNil(t, e, "Event should not be nil")
	assert.NotNil(t, e.Timestamp, "Event timestamp should not be nil")

	// Verify addressing
	addressed := e.Addressed()
	assert.False(t, addressed, "Event should not be addressed")

	e.Address = common.HexToAddress("0x456")
	addressed = e.Addressed()
	assert.True(t, addressed, "Event should be addressed")
}

func Test_EngineRelay(t *testing.T) {
	feed := make(chan core.HeuristicInput)

	eir := core.NewEngineRelay(core.PathID{}, feed)
	dummyTD := core.NewEvent(core.BlockHeader, nil)

	// Verify relay and wrapping

	go func() {
		_ = eir.RelayEvent(dummyTD)
	}()

	args := <-feed

	assert.NotNil(t, args)
	assert.Equal(t, args.PathID, core.PathID{})
	assert.Equal(t, args.Input, dummyTD)
}

func Test_SessionParams(t *testing.T) {
	isp := core.NewSessionParams(core.Layer1)
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
