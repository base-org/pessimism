package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
)

// TODO(#33): No Unit Tests for Pipeline & ETL Manager Logic

func Test_Pipeline(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() EtlStore
		testLogic         func(*testing.T, EtlStore)
	}{
		{
			name:        "Successful Add When PID Already Exists",
			function:    "addPipeline",
			description: "",

			constructionLogic: func() EtlStore {
				ctx := context.Background()

				testRegistry := newEtlStore()
				testPipeLine := getTestPipeLine(ctx)

				testRegistry.AddPipeline(core.NilPipelineUUID(), testPipeLine)
				return testRegistry
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testRouter := tc.constructionLogic()
			tc.testLogic(t, testRouter)
		})

	}
}
