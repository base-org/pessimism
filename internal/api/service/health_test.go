package service_test

import (
	"fmt"
	"testing"

	svc "github.com/base-org/pessimism/internal/api/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_GetHealth(t *testing.T) {
	ctrl := gomock.NewController(t)

	var tests = []struct {
		name        string
		description string
		function    string

		constructionLogic func() testSuite
		testLogic         func(*testing.T, testSuite)
	}{
		{
			name:        "Get Health Success",
			description: "",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := svc.Config{}
				ts := createTestSuite(ctrl, cfg)

				ts.mockEthClientInterface.EXPECT().
					HeaderByNumber(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(2)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				hc := ts.apiSvc.CheckHealth()

				assert.True(t, hc.Healthy)
				assert.True(t, hc.ChainConnectionStatus.IsL2Healthy)
				assert.True(t, hc.ChainConnectionStatus.IsL1Healthy)

			},
		},
		{
			name:        "Get Unhealthy Response",
			description: "Emulates unhealthy rpc endpoints",
			function:    "ProcessInvariantRequest",

			constructionLogic: func() testSuite {
				cfg := svc.Config{}
				ts := createTestSuite(ctrl, cfg)

				ts.mockEthClientInterface.EXPECT().
					HeaderByNumber(gomock.Any(), gomock.Any()).
					Return(nil, testErr1()).
					Times(2)

				return ts
			},

			testLogic: func(t *testing.T, ts testSuite) {
				hc := ts.apiSvc.CheckHealth()
				assert.False(t, hc.Healthy)
				assert.False(t, hc.ChainConnectionStatus.IsL2Healthy)
				assert.False(t, hc.ChainConnectionStatus.IsL1Healthy)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			testMeta := tc.constructionLogic()
			tc.testLogic(t, testMeta)
		})

	}

}
