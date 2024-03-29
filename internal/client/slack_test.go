package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

func TestSlackResponseToAlertResponse(t *testing.T) {
	testSlackSuccess := &client.SlackAPIResponse{
		Message: "ok",
		Err:     "",
	}

	testSlackFailure := &client.SlackAPIResponse{
		Message: "not ok",
		Err:     "error",
	}

	resSuc := testSlackSuccess.ToAlertResponse()
	resFail := testSlackFailure.ToAlertResponse()

	assert.Equal(t, core.SuccessStatus, resSuc.Status)
	assert.Equal(t, "", resSuc.Message)
	assert.Equal(t, core.FailureStatus, resFail.Status)
	assert.Equal(t, "error", resFail.Message)
}
