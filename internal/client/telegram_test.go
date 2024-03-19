package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/base-org/pessimism/internal/client"
	"github.com/base-org/pessimism/internal/core"
)

func TestTelegramResponseToAlertResponse(t *testing.T) {
	// Test case for a successful Telegram response
	testTelegramSuccess := &client.TelegramAPIResponse{
		Ok:     true,
		Result: nil,
		Error:  "",
	}

	// Test case for a failed Telegram response
	testTelegramFailure := &client.TelegramAPIResponse{
		Ok:    false,
		Error: "error message",
	}

	resSuc := testTelegramSuccess.ToAlertResponse()
	resFail := testTelegramFailure.ToAlertResponse()

	// Assert that the success case is correctly interpreted
	assert.Equal(t, core.SuccessStatus, resSuc.Status)
	assert.Equal(t, "Message sent successfully", resSuc.Message)

	// Assert that the failure case is correctly interpreted
	assert.Equal(t, core.FailureStatus, resFail.Status)
	assert.Equal(t, "error message", resFail.Message)
}
