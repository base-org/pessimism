package alert_test

import (
	"context"
	"testing"
	"time"

	"github.com/base-org/pessimism/internal/alert"
	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_EventLoop(t *testing.T) {
	sc := mocks.NewMockSlackClient(gomock.NewController(t))
	pdc := mocks.NewMockPagerdutyClient(gomock.NewController(t))

	ctx := context.Background()

	am := alert.NewManager(ctx, sc, pdc)

	go func() {
		_ = am.EventLoop()
	}()

	defer func() {
		_ = am.Shutdown()
	}()

	ingress := am.Transit()

	testAlert := core.Alert{
		Dest:  core.Slack,
		SUUID: core.NilSUUID(),
	}

	err := am.AddSession(core.NilSUUID(),
		&core.AlertPolicy{
			Dest: core.Slack.String(),
			Msg:  "test",
		})
	assert.Nil(t, err)

	sc.EXPECT().PostData(gomock.Any(), gomock.Any()).
		Return(&client.SlackAPIResponse{
			Ok:  true,
			Err: "",
		}, nil).
		Times(1)

	ingress <- testAlert

	time.Sleep(1 * time.Second)

	testID := core.MakeSUUID(1, 1, 1)

	testAlert = core.Alert{
		Dest:  core.ThirdParty,
		SUUID: testID,
	}

	ingress <- testAlert

	time.Sleep(1 * time.Second)
}
