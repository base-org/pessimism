//go:generate mockgen -package mocks --destination=../mocks/mock_sns.go . SNSClient

package client

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/logging"

	"go.uber.org/zap"
)

// SNSClient ... An interface for SNS clients to implement
type SNSClient interface {
	AlertClient
}

// SNSConfig ... Configuration for SNS client
type SNSConfig struct {
	TopicArn string
	Endpoint string
}

// SNSMessagePayload ... The json message payload published to SNS
type SNSMessagePayload struct {
	Network       string    `json:"network"`
	HeuristicType string    `json:"heuristic_type"`
	Severity      string    `json:"severity"`
	PathID        string    `json:"path_id"`
	HeuristicID   string    `json:"heuristic_id"`
	Timestamp     time.Time `json:"timestamp"`
	Content       string    `json:"content"`
}

// SNSMessage ... The SNS message structure. Required for SNS Publish API
type SNSMessage struct {
	Default string `json:"default"`
}

type snsClient struct {
	svc      *sns.SNS
	name     string
	topicArn string
}

// NewSNSClient ... Initializer
func NewSNSClient(cfg *SNSConfig, name string) SNSClient {
	if cfg.TopicArn == "" {
		logging.NoContext().Warn("No SNS topic ARN provided")
	}

	// Initialize a session that the SDK will use to load configuration,
	// credentials, and region. AWS_REGION, AWS_SECRET_ACCESS_KEY, and AWS_ACCESS_KEY_ID should be set in the
	// environment's runtime
	// Note: If session is to arbitrarily crash, there is a possibility that message publishing will fail
	sess, err := session.NewSession(&aws.Config{
		Endpoint: aws.String(cfg.Endpoint),
	})
	if err != nil {
		logging.NoContext().Error("Failed to create AWS session", zap.Error(err))
		return nil
	}

	return &snsClient{
		svc:      sns.New(sess),
		topicArn: cfg.TopicArn,
		name:     name,
	}
}

// Marshal ... Marshals the SNS message payload
func (p *SNSMessagePayload) Marshal() ([]byte, error) {
	payloadBytes, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}

	msg := &SNSMessage{
		Default: string(payloadBytes),
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	return msgBytes, nil
}

// PostEvent ... Publishes an event to an SNS topic ARN
func (sc snsClient) PostEvent(_ context.Context, event *AlertEventTrigger) (*AlertAPIResponse, error) {
	msgPayload, err := event.ToSNSMessagePayload().Marshal()
	if err != nil {
		return nil, err
	}

	// Publish a message to the topic
	result, err := sc.svc.Publish(&sns.PublishInput{
		Message:          aws.String(string(msgPayload)),
		MessageStructure: aws.String("json"),
		TopicArn:         &sc.topicArn,
	})
	if err != nil {
		return &AlertAPIResponse{
			Status:  core.FailureStatus,
			Message: err.Error(),
		}, err
	}

	return &AlertAPIResponse{
		Status:  core.SuccessStatus,
		Message: *result.MessageId,
	}, nil
}

func (sc snsClient) GetName() string {
	return sc.name
}
