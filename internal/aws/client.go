package aws

import "github.com/aws/aws-sdk-go-v2/service/dynamodb"

type AWSClient struct {
	client dynamodb.Client
}
