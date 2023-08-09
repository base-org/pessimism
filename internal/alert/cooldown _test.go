package alert_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_CoolDown(t *testing.T) {
	var testCases = []struct {
		name         string
		construction func() alert.CoolDownHandler
		testFunc     func(t *testing.T, cdh alert.CoolDownHandler)
	}{
		{
			name:         "Test_CoolDownHandler",
			construction: alert.NewCoolDownHandler,
			testFunc: func(t *testing.T, cdh alert.CoolDownHandler) {
				// Add a cooldown for one second
				cdh.Add(core.NilSUUID(), time.Duration(1_000_000_000))

				cooled := cdh.IsCoolDown(core.NilSUUID())
				assert.True(t, cooled)

				// Sleep for one second
				time.Sleep(1_000_000_000)
				cdh.Update()
				cooled = cdh.IsCoolDown(core.NilSUUID())
				assert.False(t, cooled)
			},
		},
	}

	for i, tc := range testCases {
		t.Run(fmt.Sprintf("%d-%s", i, tc.name), func(t *testing.T) {
			tc.testFunc(t, tc.construction())
		})
	}

}
