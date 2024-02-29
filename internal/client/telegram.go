package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"
)

type TelegramClient interface {
	AlertClient
}

// TelegramConfig holds configuration details for creating a new Telegram client.
// Token: The Bot Token provided by BotFather upon creating a new bot (https://core.telegram.org/bots/api).
// ChatID: Unique identifier for the target chat (https://core.telegram.org/constructor/channel).
type TelegramConfig struct {
	Token  string
	ChatID string
}

type telegramClient struct {
	name   string
	token  string
	chatID string
	client *http.Client
}

func NewTelegramClient(cfg *TelegramConfig, name string) (TelegramClient, error) {
	if cfg.Token == "" {
		logging.NoContext().Warn("No Telegram bot token provided")
		return nil, errors.New("No Telegram bot token was provided")
	}

	return &telegramClient{
		token:  cfg.Token,
		chatID: cfg.ChatID,
		name:   name,
		client: &http.Client{},
	}, nil
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

	// API endpoint "https://api.telegram.org/bot%s/sendMessage" is used to send messages (https://core.telegram.org/bots/api#sendmessage)
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
	return tc.name
}
