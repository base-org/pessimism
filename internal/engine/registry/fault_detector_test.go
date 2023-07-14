package registry_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/app"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine/invariant"
	"github.com/base-org/pessimism/internal/engine/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testErr() error {
	return fmt.Errorf("test error")
}

type fdTestSuite struct {
	ctrl *gomock.Controller

	mockEthClient  *mocks.MockEthClient
	mockGethClient *mocks.MockGethClient

	fd invariant.Invariant
}

func createFdTestSuite(t *testing.T) *fdTestSuite {
	ctrl := gomock.NewController(t)
	cfg := &registry.FaultDetectorCfg{
		L2OutputOracle: "0x0000000000000000000000000000000000000000",
		L2ToL1Address:  "0x0000000000000000000000000000000000000000",
	}
	ctx := context.Background()

	mockEthClient := mocks.NewMockEthClient(ctrl)
	mockGethClient := mocks.NewMockGethClient(ctrl)

	ctx = app.InitializeContext(ctx, nil, mockEthClient, mockEthClient,
		mockGethClient)

	fd, err := registry.NewFaultDetector(ctx, cfg)
	assert.Nil(t, err)

	return &fdTestSuite{
		ctrl:           ctrl,
		mockEthClient:  mockEthClient,
		mockGethClient: mockGethClient,
		fd:             fd,
	}
}

func Test_FaultDetector(t *testing.T) {
	testLog := types.Log{
		Address: common.HexToAddress("0x0000000000000000000000000000000000000000000000000000000000000000"),
		Topics: []common.Hash{
			common.HexToHash("0xa7aaf2512769da4e444e3de247be2564225c2e7a8f74cfe528e46e17d24868e2"),
			common.HexToHash("0x50c05179b9c26cb970de294e13ed114934a4e2c8631657c3189a1d17d8d32c4d"),
			common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000256"),
			common.HexToHash("0x00000000000000000000000000000000000000000000000000000000001073b8"),
		},
	}

	var tests = []struct {
		name        string
		constructor func(t *testing.T) *fdTestSuite
		testFunc    func(t *testing.T, ts *fdTestSuite)
	}{
		{
			name:        "Failure when fetching l2 block",
			constructor: createFdTestSuite,
			testFunc: func(t *testing.T, ts *fdTestSuite) {
				ts.mockEthClient.EXPECT().
					BlockByNumber(gomock.Any(), gomock.Any()).
					Return(nil, assert.AnError).
					Times(1)

				td := core.TransitData{
					Type:  core.EventLog,
					Value: testLog,
				}

				outcome, pass, err := ts.fd.Invalidate(td)
				assert.Nil(t, outcome)
				assert.False(t, pass)
				assert.Error(t, err)

			},
		},
		{
			name:        "Failure when fetching proof response",
			constructor: createFdTestSuite,
			testFunc: func(t *testing.T, ts *fdTestSuite) {
				ts.mockEthClient.EXPECT().
					BlockByNumber(gomock.Any(), gomock.Any()).
					Return(&types.Block{}, nil).
					Times(1)

				ts.mockGethClient.EXPECT().
					GetProof(gomock.Any(), gomock.Any(),
						gomock.Any(), gomock.Any()).
					Return(nil, testErr()).
					Times(1)

				td := core.TransitData{
					Type:  core.EventLog,
					Value: testLog,
				}

				outcome, pass, err := ts.fd.Invalidate(td)
				assert.Nil(t, outcome)
				assert.False(t, pass)
				assert.Error(t, err)

			},
		},
		{
			name:        "Invalidation occurs when provided an invalid proof response",
			constructor: createFdTestSuite,
			testFunc: func(t *testing.T, ts *fdTestSuite) {

				blockEnc := common.FromHex("f9030bf901fea083cafc574e1f51ba9dc0568fc617a08ea2429fb384059c972f13b19fa1c8dd55a01dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347948888f1f195afa192cfee860698584c030f4c9db1a0ef1552a40b7165c3cd773806b9e0c165b75356e0314bf0706f279c729f51e017a05fe50b260da6308036625b850b5d6ced6d0a9f814c0688bc91ffb7b7a3a54b67a0bc37d79753ad738a6dac4921e57392f145d8887476de3f783dfa7edae9283e52b90100000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008302000001832fefd8825208845506eb0780a0bd4472abb6659ebe3ee06ee4d7b72a00a9f4d001caca51342001075469aff49888a13a5a8c8f2bb1c4843b9aca00f90106f85f800a82c35094095e7baea6a6c7c4c2dfeb977efac326af552d870a801ba09bea4c4daac7c7c52e093e6a4c35dbbcf8856f1af7b059ba20253e70848d094fa08a8fae537ce25ed8cb5af9adac3f141af69bd515bd2ba031522df09b97dd72b1b8a302f8a0018080843b9aca008301e24194095e7baea6a6c7c4c2dfeb977efac326af552d878080f838f7940000000000000000000000000000000000000001e1a0000000000000000000000000000000000000000000000000000000000000000080a0fe38ca4e44a30002ac54af7cf922a6ac2ba11b7d22f548e8ecb3f51f41cb31b0a06de6a5cbae13c0c856e33acf021b51819636cfc009d39eafb9f606d546e305a8c0")

				var block *types.Block
				err := rlp.DecodeBytes(blockEnc, &block)
				assert.NoError(t, err)

				ts.mockEthClient.EXPECT().
					BlockByNumber(gomock.Any(), gomock.Any()).
					Return(block, nil).
					Times(1)

				ts.mockGethClient.EXPECT().
					GetProof(gomock.Any(), gomock.Any(),
						gomock.Any(), gomock.Any()).
					Return(&gethclient.AccountResult{
						StorageHash: common.HexToHash("0x0"),
					}, nil).
					Times(1)

				td := core.TransitData{
					Type:  core.EventLog,
					Value: testLog,
				}

				outcome, pass, err := ts.fd.Invalidate(td)
				assert.NotNil(t, outcome)
				assert.True(t, pass)
				assert.NoError(t, err)

			},
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, test.name), func(t *testing.T) {
			ts := test.constructor(t)
			test.testFunc(t, ts)

		})
	}
}
