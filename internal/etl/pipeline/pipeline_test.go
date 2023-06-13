package pipeline

import (
	"fmt"
	"testing"
)

// TODO(#33): No Unit Tests for Pipeline & ETL Manager Logic

func Test_Pipeline(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() Pipeline
		testLogic         func(t *testing.T, pl Pipeline)
	}{
		// {
		// 	name:        "Successful Add When PID Already Exists",
		// 	function:    "addPipeline",
		// 	description: "",

		// 	constructionLogic: func() Pipeline {
		// 		return getTestPipeLine(context.Background())
		// 	},

		// 	testLogic: func(t *testing.T, pl Pipeline) {
		// 		wg := sync.WaitGroup{}

		// 		err := pl.RunPipeline(&wg)
		// 		assert.NoError(t, err)

		// 		err = pl.Close()
		// 		assert.NoError(t, err)

		// 	},
		// },
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testPipeline := tc.constructionLogic()
			tc.testLogic(t, testPipeline)
		})

	}
}
