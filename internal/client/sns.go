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
	"os"
)

// SNSClient ... An interface for SNS clients to implement
type SNSClient interface {
	AlertClient
}

// SNSConfig ... Configuration for SNS client
type SNSConfig struct {
	TopicArn string
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

	logging.NoContext().Debug("AWS Region", zap.String("region", os.Getenv("AWS_REGION")))

	// Initialize a session that the SDK will use
	sess, err := session.NewSession()
	if err != nil {
		logging.NoContext().Error("Failed to create SNS session", zap.Error(err))
		return nil
	}

	return &snsClient{
		svc:      sns.New(sess),
		topicArn: cfg.TopicArn,
		name:     name,
	}
}

// PostEvent ... Posts an event to an SNS topic ARN
func (sc snsClient) PostEvent(ctx context.Context, event *AlertEventTrigger) (*AlertAPIResponse, error) {
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
