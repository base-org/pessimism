package pipeline

// TODO - Clean up and better testing labels

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/conduit/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_Pipe_OPBlockToTransactions(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	// TODO - Make this something constant
	ts := time.Now()

	tranformFunc := func(td models.TransitData) (*models.TransitData, error) {
		parsedBlock, success := td.Value.(types.Block)
		if !success {
			return nil, fmt.Errorf("Could not parse transit value to Geth block")
		}

		txs := parsedBlock.Transactions()
		println(len(txs))

		tfTd := &models.TransitData{
			Timestamp: ts,
			Type:      "GETH.BLOCK.TRANSACTIONS",
			Value:     txs,
		}

		log.Printf("%+v", tfTd)
		return tfTd, nil
	}
	testID := uuid.New()
	outputChan := make(chan models.TransitData)
	inputChan := make(chan models.TransitData)

	router := NewOutputRouter(
		WithDirective(testID, outputChan),
	)

	pipe := NewPipe(ctx, tranformFunc, inputChan, WithRouter(router))
	blockEnc := common.FromHex("f9030bf901fea083cafc574e1f51ba9dc0568fc617a08ea2429fb384059c972f13b19fa1c8dd55a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a05fe50b260da6308036625b850b5d6ced6d0a9f814c0688bc91ffb7b7a3a54b67a0bc37d79753ad738a6dac4921e57392f145d8887476de3f783dfa7edae9283e52b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000001832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4843b9aca00f90106f85f800a82c35094095e7baea6a6c7c4c2dfeb977efac326af552d870a801ba09bea4c4daac7c7c52e093e6a4c35dbbcf8856f1af7b059ba20253e70848d094fa08a8fae537ce25ed8cb5af9adac3f141af69bd515bd2ba031522df09b97dd72b1b8a302f8a0018080843b9aca008301e24194095e7baea6a6c7c4c2dfeb977efac326af552d878080f838f7940000000000000000000000000000000000000001e1a0000000000000000000000000000000000000000000000000000000000000000080a0fe38ca4e44a30002ac54af7cf922a6ac2ba11b7d22f548e8ecb3f51f41cb31b0a06de6a5cbae13c0c856e33acf021b51819636cfc009d39eafb9f606d546e305a8c0")
	var block types.Block
	if err := rlp.DecodeBytes(blockEnc, &block); err != nil {
		t.Fatal("decode error: ", err)
	}

	count := len(block.Transactions())

	ticker := time.NewTicker(500 * time.Millisecond)

	go pipe.EventLoop()

	wg := sync.WaitGroup{}

	inputData := models.TransitData{
		Timestamp: ts,

		Type:  "GETH.BLOCK",
		Value: block,
	}
	var outputData *models.TransitData = nil

	wg.Add(1)
	go func() {
	read:
		for {
			select {
			case od := <-outputChan:
				outputData = &od
				break read

			case <-ticker.C:
				break read
			}
		}
		wg.Done()

	}()

	inputChan <- inputData

	println("Waiting for output")
	wg.Wait()

	assert.NotNil(t, outputData)

	actualTxs, success := outputData.Value.(types.Transactions)
	assert.True(t, success)
	assert.Equal(t, outputData.Timestamp, ts)

	assert.True(t, len(actualTxs) == count, fmt.Sprintf("Got %d txs, expected %d", len(actualTxs), count))

}
