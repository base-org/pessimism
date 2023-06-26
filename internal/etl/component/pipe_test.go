package component_test

import (
	"context"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/stretchr/testify/assert"
)

func Test_Pipe_Event_Flow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ts := time.Date(1969, time.April, 1, 4, 20, 0, 0, time.Local)

	// Setup component dependencies
	testID := core.MakeCUUID(6, 9, 6, 9)

	outputChan := make(chan core.TransitData)

	// Construct test component
	testPipe, err := mocks.NewDummyPipe(ctx, core.GethBlock, core.EventLog)
	assert.NoError(t, err)

	err = testPipe.AddEgress(testID, outputChan)
	assert.NoError(t, err)

	// Encoded value taken from https://github.com/ethereum/go-ethereum/blob/master/core/types/block_test.go#L36
	blockEnc := common.FromHex("f9030bf901fea083cafc574e1f51ba9dc0568fc617a08ea2429fb384059c972f13b19fa1c8dd55a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a05fe50b260da6308036625b850b5d6ced6d0a9f814c0688bc91ffb7b7a3a54b67a0bc37d79753ad738a6dac4921e57392f145d8887476de3f783dfa7edae9283e52b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000001832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4843b9aca00f90106f85f800a82c35094095e7baea6a6c7c4c2dfeb977efac326af552d870a801ba09bea4c4daac7c7c52e093e6a4c35dbbcf8856f1af7b059ba20253e70848d094fa08a8fae537ce25ed8cb5af9adac3f141af69bd515bd2ba031522df09b97dd72b1b8a302f8a0018080843b9aca008301e24194095e7baea6a6c7c4c2dfeb977efac326af552d878080f838f7940000000000000000000000000000000000000001e1a0000000000000000000000000000000000000000000000000000000000000000080a0fe38ca4e44a30002ac54af7cf922a6ac2ba11b7d22f548e8ecb3f51f41cb31b0a06de6a5cbae13c0c856e33acf021b51819636cfc009d39eafb9f606d546e305a8c0")

	var block types.Block
	err = rlp.DecodeBytes(blockEnc, &block)
	assert.NoError(t, err)

	// Start component event loop on separate go routine
	go func() {
		if err := testPipe.EventLoop(); err != nil {
			log.Printf("Got error from testPipe event loop %s", err.Error())
		}
	}()

	wg := sync.WaitGroup{}

	inputData := core.TransitData{
		Timestamp: ts,
		Type:      core.GethBlock,
		Value:     block,
	}
	var outputData core.TransitData

	// Spawn listener routine that reads for output from testPipe
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Read first value from channel and return
		for output := range outputChan {
			outputData = output
			return
		}

	}()

	entryChan, err := testPipe.GetIngress(core.GethBlock)
	assert.NoError(t, err)

	entryChan <- inputData

	// Wait for pipe to transform block data into a transaction slice
	wg.Wait()

	assert.NotNil(t, outputData)

	_, success := outputData.Value.(types.Block)
	assert.True(t, success)
	assert.Equal(t, outputData.Timestamp, ts, "Timestamp failed to verify")

}
