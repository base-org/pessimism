package oracle

import (
	"context"
	"fmt"
	"math/big"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/trie"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_GetCurrentHeightFromNetwork(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	logging.NewLogger(nil, "development")
	defer cancel()

	// setup mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testObj := mocks.NewMockEthClientInterface(ctrl)

	header := types.Header{
		ParentHash: common.HexToHash("0x123456789"),
		Number:     big.NewInt(5),
	}
	// setup expectations
	testObj.
		EXPECT().
		HeaderByNumber(gomock.Any(), gomock.Any()).
		Return(&header, nil)

	od := &GethBlockODef{cfg: &core.ClientConfig{
		NumOfRetries: 3,
	}, currHeight: nil, client: testObj}

	assert.Equal(t, od.getCurrentHeightFromNetwork(ctx).Number, header.Number)
}

func Test_GetHeightToProcess(t *testing.T) {

	ctx, cancel := context.WithCancel(context.Background())
	logging.NewLogger(nil, "development")
	defer cancel()

	// setup mock
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	testObj := mocks.NewMockEthClientInterface(ctrl)

	header := types.Header{
		ParentHash: common.HexToHash("0x123456789"),
		Number:     big.NewInt(5),
	}
	testObj.
		EXPECT().
		HeaderByNumber(gomock.Any(), gomock.Any()).
		Return(&header, nil).
		AnyTimes()

	od := &GethBlockODef{cfg: &core.ClientConfig{
		NumOfRetries: 3,
	}, currHeight: big.NewInt(123), client: testObj}

	assert.Equal(t, od.getHeightToProcess(ctx), big.NewInt(123))

	od.currHeight = nil
	od.cfg.StartHeight = big.NewInt(123)
	assert.Equal(t, od.getHeightToProcess(ctx), big.NewInt(123))

	od.currHeight = nil
	od.cfg.StartHeight = nil
	assert.Nil(t, od.getHeightToProcess(ctx))
}

func Test_Backroutine(t *testing.T) {
	logging.NewLogger(nil, "development")
	var tests = []struct {
		name        string
		description string

		constructionLogic func() (*GethBlockODef, chan core.TransitData)
		testLogic         func(*testing.T, *GethBlockODef, chan core.TransitData)
	}{

		{
			name:        "Current network height check",
			description: "Check if network height check is less than starting height",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				header := types.Header{
					ParentHash: common.HexToHash("0x123456789"),
					Number:     big.NewInt(5),
				}
				// setup expectationss
				testObj.
					EXPECT().
					HeaderByNumber(gomock.Any(), gomock.Any()).
					Return(&header, nil).
					AnyTimes()

				od := &GethBlockODef{cfg: &core.ClientConfig{
					NumOfRetries: 3,
				}, currHeight: nil, client: testObj}

				outChan := make(chan core.TransitData)

				return od, outChan
			},

			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.BackTestRoutine(ctx, outChan, big.NewInt(7), big.NewInt(10))
				assert.Error(t, err)
				assert.EqualError(t, err, "start height cannot be more than the latest height from network")
			},
		},
		{
			name:        "Successful Height check",
			description: "Ending height cannot be less than the Starting height",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				od := &GethBlockODef{cfg: &core.ClientConfig{
					NumOfRetries: 3,
				}, currHeight: nil, client: testObj}

				outChan := make(chan core.TransitData)

				return od, outChan
			},

			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.BackTestRoutine(ctx, outChan, big.NewInt(2), big.NewInt(1))
				assert.Error(t, err)
				assert.EqualError(t, err, "start height cannot be more than the end height")
			},
		},
		// Leaving this here to help devs test infinite loops
		//
		//{
		//	name:        "Header fetch retry exceeded error check",
		//	description: "Check if the header fetch retry fails after 3 retries, total 4 tries.",
		//
		//	constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
		//		testObj := new(EthClientMocked)
		//
		//		// setup expectations
		//		testObj.On("DialContext", mock.Anything, "pass test").Return(nil)
		//		testObj.On("HeaderByNumber", mock.Anything, mock.Anything).Return(nil, errors.New("no header for you"))
		//
		//		od := &GethBlockODef{cfg: &core.ClientConfig{
		//			RPCEndpoint:  "pass test",
		//			NumOfRetries: 3,
		//		}, currHeight: nil, client: testObj}
		//
		//		outChan := make(chan core.TransitData)
		//		return od, outChan
		//	},
		//
		//	testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {
		//
		//		ctx, cancel := context.WithCancel(context.Background())
		//		defer cancel()
		//
		//		err := od.BackTestRoutine(ctx, outChan, big.NewInt(1), big.NewInt(2))
		//		assert.Error(t, err)
		//		assert.EqualError(t, err, "no header for you")
		//	},
		// },
		{
			name:        "Backroutine happy path test",
			description: "Backroutine works and channel should have 4 messages waiting.",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				header := types.Header{
					ParentHash: common.HexToHash("0x123456789"),
					Number:     big.NewInt(7),
				}
				block := types.NewBlock(&header, nil, nil, nil, trie.NewStackTrie(nil))
				// setup expectations
				testObj.
					EXPECT().
					HeaderByNumber(gomock.Any(), gomock.Any()).
					Return(&header, nil).
					AnyTimes()
				testObj.
					EXPECT().
					BlockByNumber(gomock.Any(), gomock.Any()).
					Return(block, nil).
					AnyTimes()

				od := &GethBlockODef{cfg: &core.ClientConfig{
					NumOfRetries: 3,
					PollInterval: 1000,
				}, currHeight: nil, client: testObj}

				outChan := make(chan core.TransitData, 2)

				return od, outChan
			},

			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.BackTestRoutine(ctx, outChan, big.NewInt(5), big.NewInt(6))
				assert.NoError(t, err)
				close(outChan)

				for m := range outChan {
					val := m.Value.(types.Block) //nolint:errcheck // converting to type from any for getting internal values
					assert.Equal(t, val.ParentHash(), common.HexToHash("0x123456789"))
				}
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			od, outChan := tc.constructionLogic()
			tc.testLogic(t, od, outChan)
		})

	}
}

func Test_ReadRoutine(t *testing.T) {
	logging.NewLogger(nil, "development")
	var tests = []struct {
		name        string
		description string

		constructionLogic func() (*GethBlockODef, chan core.TransitData)
		testLogic         func(*testing.T, *GethBlockODef, chan core.TransitData)
	}{

		{
			name:        "Successful Height check 1",
			description: "Ending height cannot be less than the Starting height",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				od := &GethBlockODef{cfg: &core.ClientConfig{
					StartHeight:  big.NewInt(2),
					EndHeight:    big.NewInt(1),
					NumOfRetries: 3,
					PollInterval: 1000,
				}, currHeight: nil, client: testObj}
				outChan := make(chan core.TransitData)
				return od, outChan
			},
			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.ReadRoutine(ctx, outChan)
				assert.Error(t, err)
				assert.EqualError(t, err, "start height cannot be more than the end height")
			},
		},
		{
			name:        "Successful Height check 2",
			description: "Cannot have start height nil, i.e, latest block and end height configured",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				od := &GethBlockODef{cfg: &core.ClientConfig{
					StartHeight:  nil,
					EndHeight:    big.NewInt(1),
					NumOfRetries: 3,
					PollInterval: 1000,
				}, currHeight: nil, client: testObj}
				outChan := make(chan core.TransitData)
				return od, outChan
			},

			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.ReadRoutine(ctx, outChan)
				assert.Error(t, err)
				assert.EqualError(t, err, "cannot start with latest block height with end height configured")
			},
		},
		{
			description: "Making sure that number of blocks fetched matches the assumption. Number of messages should be 5, in the channel",

			constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
				// setup mock
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				testObj := mocks.NewMockEthClientInterface(ctrl)

				header := types.Header{
					ParentHash: common.HexToHash("0x123456789"),
					Number:     big.NewInt(7),
				}
				block := types.NewBlock(&header, nil, nil, nil, trie.NewStackTrie(nil))

				testObj.
					EXPECT().
					HeaderByNumber(gomock.Any(), gomock.Any()).
					Return(&header, nil).
					AnyTimes()
				testObj.
					EXPECT().
					BlockByNumber(gomock.Any(), gomock.Any()).
					Return(block, nil).
					AnyTimes()

				od := &GethBlockODef{cfg: &core.ClientConfig{
					StartHeight:  big.NewInt(1),
					EndHeight:    big.NewInt(5),
					NumOfRetries: 3,
					PollInterval: 1000,
				}, currHeight: nil, client: testObj}
				outChan := make(chan core.TransitData, 10)
				return od, outChan
			},

			testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {

				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				err := od.ReadRoutine(ctx, outChan)
				assert.NoError(t, err)
				close(outChan)
				assert.Equal(t, len(outChan), 5)
			},
		},
		// Leaving this here to help devs test infinite loops
		//
		//{
		//	name:        "Latest block check",
		//	description: "Making sure that number of blocks fetched matches the assumption. Number of messages should be 5, in the channel",
		//
		//	constructionLogic: func() (*GethBlockODef, chan core.TransitData) {
		//		testObj := new(EthClientMocked)
		//		header := types.Header{
		//			ParentHash: common.HexToHash("0x123456789"),
		//			Number:     big.NewInt(1),
		//		}
		//		block := types.NewBlock(&header, nil, nil, nil, trie.NewStackTrie(nil))
		//		// setup expectations
		//		testObj.On("DialContext", mock.Anything, "pass test").Return(nil)
		//		testObj.On("HeaderByNumber", mock.Anything, mock.Anything).Return(&header, nil)
		//		testObj.On("BlockByNumber", mock.Anything, mock.Anything).Return(block, nil)
		//
		//		od := &GethBlockODef{cfg: &core.ClientConfig{
		//			RPCEndpoint:  "pass test",
		//			StartHeight:  nil,
		//			EndHeight:    nil,
		//			NumOfRetries: 3,
		//			PollInterval: 1000,

		//		}, currHeight: nil, client: testObj}
		//		outChan := make(chan core.TransitData, 10)
		//		return od, outChan
		//	},
		//
		//	testLogic: func(t *testing.T, od *GethBlockODef, outChan chan core.TransitData) {
		//
		//		ctx, cancel := context.WithCancel(context.Background())
		//		defer cancel()
		//
		//		err := od.ReadRoutine(ctx, outChan)
		//		assert.NoError(t, err)
		//		close(outChan)
		//		assert.Equal(t, len(outChan), 5)
		//	},
		// },
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			od, outChan := tc.constructionLogic()
			tc.testLogic(t, od, outChan)
		})

	}
}
