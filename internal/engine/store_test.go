package engine_test

import (
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/engine"
	"github.com/stretchr/testify/assert"
)

func TestSessionStore(t *testing.T) {
	// Setup
	ss := engine.NewSessionStore()

	// Test GetInvSessionByUUID
	_, err := ss.GetInvSessionByUUID(core.NilSUUID())
	assert.Error(t, err, "should error")
}
