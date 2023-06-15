package pipe_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry/pipe"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/state"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

const (
	niceAddr = "0x0000000000000000000000000000000000000069"
	blansfer = "blansfer(address,address,uint256)"
)

type testSuite struct {
	ctx        context.Context
	ctrl       *gomock.Controller
	mockClient *mocks.MockEthClient
	ed         *pipe.EventDefinition
}

func createTestSuite(t *testing.T) *testSuite {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	mockClient := mocks.NewMockEthClient(ctrl)
	cc := &core.ClientConfig{}

	ed := pipe.NewEventDefinition(cc, mockClient)

	return &testSuite{ctx, ctrl, mockClient, ed}
}

func Test_ComponentRegistry(t *testing.T) {
	var tests = []struct {
		name        string
		description string

		constructionLogic func() *testSuite
		testLogic         func(*testing.T, *testSuite)
	}{
		{
			name:        "Successful Block to Event Transform Flow",
			description: "TBD",

			constructionLogic: func() *testSuite {
				ts := createTestSuite(t)

				// Setup state
				ss := state.NewMemState()
				// TODO(#69): State Key Representation is Insecure
				sk := state.MakeKey(core.EventLog, core.AddressKey, true)
				_, err := ss.SetSlice(ts.ctx, sk.WithPUUID(core.NilPipelineUUID()), niceAddr)
				if err != nil {
					panic(err)
				}

				nestedKey := state.MakeKey(core.EventLog, niceAddr, true).WithPUUID(core.NilPipelineUUID())

				_, err = ss.SetSlice(ts.ctx, nestedKey, blansfer)
				if err != nil {
					panic(err)
				}

				ts.ctx = context.WithValue(ts.ctx, state.Default, ss)

				hash := crypto.Keccak256Hash([]byte(blansfer))
				// Setup mock client

				retValue := []types.Log{{
					Address: common.HexToAddress(niceAddr),
					Topics:  []common.Hash{hash},
				}}

				ts.mockClient.EXPECT().
					DialContext(gomock.Any(), gomock.Any()).
					Return(nil)

				ts.mockClient.EXPECT().
					FilterLogs(gomock.Any(), gomock.Any()).
					Return(retValue, nil)

				return ts
			},
			testLogic: func(t *testing.T, ts *testSuite) {
				// Configure
				err := ts.ed.ConfigureRoutine(core.NilPipelineUUID())
				assert.NoError(t, err)

				blockEnc := common.FromHex("f9030bf901fea083cafc574e1f51ba9dc0568fc617a08ea2429fb384059c972f13b19fa1c8dd55a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a05fe50b260da6308036625b850b5d6ced6d0a9f814c0688bc91ffb7b7a3a54b67a0bc37d79753ad738a6dac4921e57392f145d8887476de3f783dfa7edae9283e52b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000001832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4843b9aca00f90106f85f800a82c35094095e7baea6a6c7c4c2dfeb977efac326af552d870a801ba09bea4c4daac7c7c52e093e6a4c35dbbcf8856f1af7b059ba20253e70848d094fa08a8fae537ce25ed8cb5af9adac3f141af69bd515bd2ba031522df09b97dd72b1b8a302f8a0018080843b9aca008301e24194095e7baea6a6c7c4c2dfeb977efac326af552d878080f838f7940000000000000000000000000000000000000001e1a0000000000000000000000000000000000000000000000000000000000000000080a0fe38ca4e44a30002ac54af7cf922a6ac2ba11b7d22f548e8ecb3f51f41cb31b0a06de6a5cbae13c0c856e33acf021b51819636cfc009d39eafb9f606d546e305a8c0")

				var block types.Block
				err = rlp.DecodeBytes(blockEnc, &block)
				assert.NoError(t, err)

				// Run
				td := core.TransitData{
					Type:  core.GethBlock,
					Value: block,
				}

				ret, err := ts.ed.Transform(ts.ctx, td)
				assert.NoError(t, err)
				assert.NotNil(t, ret)

				assert.Greater(t, len(ret), 0)

				// Validate
				assert.Equal(t, core.EventLog, ret[0].Type)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ts := test.constructionLogic()
			test.testLogic(t, ts)
		})
	}

}
