#!/bin/sh

echo "Initializing localstack SNS topic..."

awslocal sns create-topic --name multi-directive-test-topic
awslocal sns create-topic --name alert-cooldown-test-topic
awslocal sqs create-queue --queue-name multi-directive-test-queue
awslocal sqs create-queue --queue-name alert-cooldown-test-queue
awslocal sns subscribe --topic-arn "arn:aws:sns:us-east-1:000000000000:multi-directive-test-topic" --protocol sqs --notification-endpoint "arn:aws:sqs:us-east-1:000000000000:multi-directive-test-queue"
awslocal sns subscribe --topic-arn "arn:aws:sns:us-east-1:000000000000:alert-cooldown-test-topic" --protocol sqs --notification-endpoint "arn:aws:sqs:us-east-1:000000000000:alert-cooldown-test-queue"