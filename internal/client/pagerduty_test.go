package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
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
