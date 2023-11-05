package etl

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/registry"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/state"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestETL(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() ETL
		testLogic         func(t *testing.T, m ETL)
	}{
		{
			name:     "Success - Subscription Process",
			function: "CreateProcess",

			constructionLogic: func() ETL {
				r := registry.New()
				ctrl := gomock.NewController(t)

				ctx, _ := mocks.Context(context.Background(), ctrl)

				ctx = context.WithValue(ctx, core.State, state.NewMemState())

				return New(ctx, NewAnalyzer(r), r, NewEtlStore(), NewGraph(), nil)
			},

			testLogic: func(t *testing.T, etl ETL) {
				cUUID := core.MakeProcessID(1, 1, 1, 1)

				register, err := registry.New().GetDataTopic(core.BlockHeader)

				assert.NoError(t, err)

				cc := &core.ClientConfig{
					Network: core.Layer1,
				}
				p, err := etl.CreateProcess(cc, cUUID, core.PathID{}, register)
				assert.NoError(t, err)

				assert.Equal(t, p.ID(), cUUID)
				assert.Equal(t, p.Type(), register.ProcessType)
				assert.Equal(t, p.EmitType(), register.DataType)

			},
		},
		{
			name:        "Successful Path Creation",
			function:    "CreatePath",
			description: "CreatePath should reuse an existing path when necessary",

			constructionLogic: func() ETL {
				reg := registry.New()
				ctrl := gomock.NewController(t)

				ctx, ms := mocks.Context(context.Background(), ctrl)

				ms.MockL1Node.EXPECT().BlockHeaderByNumber(gomock.Any()).Return(nil, fmt.Errorf("keep going")).AnyTimes()

				ctx = context.WithValue(ctx, core.State, state.NewMemState())

				return New(ctx, NewAnalyzer(reg), reg, NewEtlStore(), NewGraph(), nil)
			},

			testLogic: func(t *testing.T, etl ETL) {
				pCfg := &core.PathConfig{
					Network:  core.Layer1,
					DataType: core.Log,
					PathType: core.Live,
					ClientConfig: &core.ClientConfig{
						Network:      core.Layer1,
						PollInterval: time.Hour * 1,
					},
				}

				id1, reuse, err := etl.CreateProcessPath(pCfg)
				assert.NoError(t, err)
				assert.False(t, reuse)
				assert.NotEqual(t, id1, core.PathID{})

				// Now create a new path with the same config
				// & ensure that the previous path is reused

				id2, reuse, err := etl.CreateProcessPath(pCfg)
				assert.NoError(t, err)
				assert.True(t, reuse)
				assert.Equal(t, id1, id2)

				// Now run the path
				err = etl.Run(id1)
				assert.NoError(t, err)

				// Ensure shutdown works
				go func() {
					_ = etl.EventLoop()
				}()
				err = etl.Shutdown()
				assert.NoError(t, err)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testPath := tc.constructionLogic()
			tc.testLogic(t, testPath)
		})

	}

}
