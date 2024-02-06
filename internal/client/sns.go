//go:generate mockgen -package mocks --destination=../mocks/mock_sns.go . SNSClient

package client

import (
	"context"

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

// PostEvent ... Posts an event to an SNS topic ARN
func (sc snsClient) PostEvent(_ context.Context, event *AlertEventTrigger) (*AlertAPIResponse, error) {
	// Publish a message to the topic
	result, err := sc.svc.Publish(&sns.PublishInput{
		MessageAttributes: getAttributesFromEvent(event),
		Message:           &event.Message,
		TopicArn:          &sc.topicArn,
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

// getAttributesFromEvent ... Helper method to get attributes from an AlertEventTrigger
func getAttributesFromEvent(event *AlertEventTrigger) map[string]*sns.MessageAttributeValue {
	return map[string]*sns.MessageAttributeValue{
		"severity": {
			DataType:    aws.String("String"),
			StringValue: aws.String(event.Severity.String()),
		},
		"dedup_key": {
			DataType:    aws.String("String"),
			StringValue: aws.String(event.DedupKey.String()),
		},
	}
}

func (sc snsClient) GetName() string {
	return sc.name
}
