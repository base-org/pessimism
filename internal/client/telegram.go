package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

type TelegramClient interface {
	AlertClient
}

type TelegramConfig struct {
	Token  string
	ChatID string
}

type telegramClient struct {
	token  string
	chatID string
	client *http.Client
}

func NewTelegramClient(cfg *TelegramConfig) TelegramClient {
	if cfg.Token == "" {
		logging.NoContext().Warn("No Telegram token provided")
	}

	return &telegramClient{
		token:  cfg.Token,
		chatID: cfg.ChatID,
		client: &http.Client{},
	}
}

type TelegramPayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

type TelegramAPIResponse struct {
	Ok     bool            `json:"ok"`
	Result json.RawMessage `json:"result"` // Might not be needed for basic response handling
	Error  string          `json:"description"`
}

func (tr *TelegramAPIResponse) ToAlertResponse() *AlertAPIResponse {
	if tr.Ok {
		return &AlertAPIResponse{
			Status:  core.SuccessStatus,
			Message: "Message sent successfully",
		}
	}
	return &AlertAPIResponse{
		Status:  core.FailureStatus,
		Message: tr.Error,
	}
}

func (tc *telegramClient) PostEvent(ctx context.Context, data *AlertEventTrigger) (*AlertAPIResponse, error) {
	payload := TelegramPayload{
		ChatID: tc.chatID,
		Text:   data.Message,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", tc.token)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp TelegramAPIResponse
	if err = json.Unmarshal(respBytes, &apiResp); err != nil {
		return nil, fmt.Errorf("could not unmarshal telegram response: %w", err)
	}

	return apiResp.ToAlertResponse(), nil
}

func (tc *telegramClient) GetName() string {
	return "TelegramClient"
}
