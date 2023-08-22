package client_test

import (
	"github.com/base-org/pessimism/internal/core"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/base-org/pessimism/internal/client"
)

func Test_PagerdutyAPIResponse_To_AlertAPIResponse(t *testing.T) {

	testPdSuccess := &client.PagerDutyAPIResponse{
		Status:  "success",
		Message: "test",
	}

	testPdFailure := &client.PagerDutyAPIResponse{
		Status:  "failure",
		Message: "test",
	}

	resSuc := testPdSuccess.ToAlertResponse()
	resFail := testPdFailure.ToAlertResponse()

	assert.Equal(t, resSuc.Status, core.SuccessStatus)
	assert.Equal(t, resFail.Status, core.FailureStatus)
}
