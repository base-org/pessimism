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
				cdh.Add(core.NilSUUID(), time.Duration(1))

				cooled := cdh.IsCoolDown(core.NilSUUID())
				assert.True(t, cooled)

				time.Sleep(time.Second * 1)
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
