package e2e_test

import "testing"

func TestBackfill(t *testing.T) {

	// 1 - Move the L1 chain forward 10 blocks and create system transaction

	// 2 - Move the L1 chain forward another 10 blocks

	// 3 - Wire dessimism to start backfilling a system tx heuristic from height 0

	// 4 - Assert that pessimism detected the system tx event
}
