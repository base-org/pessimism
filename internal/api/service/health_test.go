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

				return createTestSuite(ctrl, cfg)
			},

			testLogic: func(t *testing.T, ts testSuite) {
				hc := ts.apiSvc.CheckHealth()

				assert.True(t, hc.Healthy)
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
