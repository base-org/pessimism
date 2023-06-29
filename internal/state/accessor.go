package state

import (
	"context"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

// InsertUnique ... Inserts a new unique entry into the state store
// No error if the entry already exists, only warning
// NOTE: Loading state from context is a temporary solution
func InsertUnique(ctx context.Context, sk *core.StateKey, value string) error {
	ss, err := FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = ss.SetSlice(ctx, sk, value)

	if err != nil && !isValAlreadySetError(err) {
		return err
	}

	if err != nil {
		logging.WithContext(ctx).Warn("Invariant session already exists in state store")
	}
	return nil
}
