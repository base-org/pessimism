package oracle

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/config"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/base-org/pessimism/internal/state"
	gomock "github.com/golang/mock/gomock"
)

type testSuit struct {
}

func Test_Balance_Oracle_ReadRoutine(t *testing.T) {

	ctx := context.Background()
	localState := state.NewMemState()
	ctx = context.WithValue(ctx, "state", localState)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockClient := mocks.NewMockEthClientInterface(ctrl)
	mockClient.EXPECT().BalanceAt()

	testDef := NewAddressBalanceODef(&config.OracleConfig{}, mockClient, nil)

	compChan := make(chan core.TransitData)

	go func() {
		testDef.ReadRoutine(ctx, compChan)
	}()

}
