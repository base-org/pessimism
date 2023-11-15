package etl_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl"

	"github.com/base-org/pessimism/internal/etl/process"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/stretchr/testify/assert"
)

var (
	nilID = core.MakePathID(0, core.ProcessID{}, core.ProcessID{})

	cID1 = core.MakeProcessID(0, 0, 0, 0)
	cID2 = core.MakeProcessID(0, 0, 0, 0)
)

func getTestPath(ctx context.Context) etl.Path {

	p1, err := mocks.NewReader(ctx, core.BlockHeader, process.WithID(cID2))
	if err != nil {
		panic(err)
	}

	p2, err := mocks.NewSubscriber(ctx, core.BlockHeader, core.Log, process.WithID(cID1))
	if err != nil {
		panic(err)
	}

	procs := []process.Process{
		p1,
		p2,
	}

	path, err := etl.NewPath(&core.PathConfig{}, core.PathID{}, procs)
	if err != nil {
		panic(err)
	}

	return path
}

func TestStore(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() *etl.Store
		testLogic         func(*testing.T, *etl.Store)
	}{
		{
			name:        "Successful Add When PID Already Exists",
			function:    "AddPath",
			description: "",

			constructionLogic: func() *etl.Store {
				ctx := context.Background()

				store := etl.NewStore()
				testPath := getTestPath(ctx)

				store.AddPath(core.PathID{}, testPath)
				return store
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				ctx := context.Background()
				testPath := getTestPath(ctx)

				pID2 := core.MakePathID(
					0,
					core.MakeProcessID(0, 0, 0, 1),
					core.MakeProcessID(0, 0, 0, 1),
				)

				store.AddPath(pID2, testPath)

				for _, p := range testPath.Processes() {
					ids, err := store.GetPathIDs(p.ID())

					assert.NoError(t, err)
					assert.Len(t, ids, 2)
					assert.Equal(t, ids[0].ID, nilID.ID)
					assert.Equal(t, ids[1].ID, pID2.ID)
				}

			},
		},
		{
			name:        "Successful Add When PID Does Not Exists",
			function:    "AddPath",
			description: "",

			constructionLogic: func() *etl.Store {
				pr := etl.NewStore()
				return pr
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				ctx := context.Background()
				testPath := getTestPath(ctx)

				pID := core.MakePathID(
					0,
					core.MakeProcessID(0, 0, 0, 1),
					core.MakeProcessID(0, 0, 0, 1),
				)

				store.AddPath(pID, testPath)

				for _, p := range testPath.Processes() {
					ids, err := store.GetPathIDs(p.ID())

					assert.NoError(t, err)
					assert.Len(t, ids, 1)
					assert.Equal(t, ids[0], pID)
				}

			},
		},
		{
			name:        "Successful Retrieval When CID Is Valid Key",
			function:    "getpathIDs",
			description: "",

			constructionLogic: etl.NewStore,
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				store.Link(cID, pID)

				expectedIDs := []core.PathID{pID}
				actualIDs, err := store.GetPathIDs(cID)

				assert.NoError(t, err)
				assert.Equal(t, expectedIDs, actualIDs)

			},
		},
		{
			name:        "Failed Retrieval When CID Is Invalid Key",
			function:    "getpathIDs",
			description: "",

			constructionLogic: etl.NewStore,
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)

				_, err := store.GetPathIDs(cID)

				assert.Error(t, err)
			},
		},
		{
			name:        "Failed Retrieval When PID Is Invalid Key",
			function:    "getpath",
			description: "",

			constructionLogic: etl.NewStore,
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				_, err := store.GetPathByID(pID)
				assert.Error(t, err)

			},
		}, {
			name:        "Failed Retrieval When Matching UUID Cannot Be Found",
			function:    "getpath",
			description: "",

			constructionLogic: func() *etl.Store {
				store := etl.NewStore()
				return store
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				pLine := getTestPath(context.Background())

				store.AddPath(pID, pLine)

				pID2 := core.MakePathID(0, cID, cID)
				_, err := store.GetPathByID(pID2)

				assert.Error(t, err)

			},
		}, {
			name:        "Successful Retrieval",
			function:    "getpath",
			description: "",

			constructionLogic: func() *etl.Store {
				store := etl.NewStore()
				return store
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				expected := getTestPath(context.Background())

				store.AddPath(pID, expected)

				actualPline, err := store.GetPathByID(pID)

				assert.NoError(t, err)
				assert.Equal(t, expected, actualPline)
			},
		},
		{
			name:        "Successful path Fetch",
			function:    "Paths",
			description: "",

			constructionLogic: func() *etl.Store {
				store := etl.NewStore()
				return store
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				expected := getTestPath(context.Background())

				store.AddPath(pID, expected)

				paths := store.Paths()

				assert.Len(t, paths, 1)
				assert.Equal(t, paths[0], expected)
			},
		},
		{
			name:        "Successful Active Count Call",
			function:    "ActiveCount",
			description: "",

			constructionLogic: func() *etl.Store {
				store := etl.NewStore()
				return store
			},
			testLogic: func(t *testing.T, store *etl.Store) {
				cID := core.MakeProcessID(0, 0, 0, 0)
				pID := core.MakePathID(0, cID, cID)

				expected := getTestPath(context.Background())

				store.AddPath(pID, expected)

				count := store.ActiveCount()
				assert.Equal(t, count, 0)
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
