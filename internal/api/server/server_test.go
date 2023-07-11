package server_test

import (
	"context"
	"testing"

	"github.com/base-org/pessimism/internal/api/handlers"
	"github.com/base-org/pessimism/internal/api/server"
	"github.com/base-org/pessimism/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_ServerFlow(t *testing.T) {
	cfg := &server.Config{
		Host: "localhost",
		Port: 8080,
	}

	mockSvc := mocks.NewMockService(gomock.NewController(t))

	handlers, err := handlers.New(context.Background(), mockSvc)
	assert.NoError(t, err)

	svr, shutdown, err := server.New(context.Background(), cfg, handlers)
	assert.NoError(t, err)

	assert.NotNil(t, svr)
	assert.NotNil(t, shutdown)

	shutdown()
}
