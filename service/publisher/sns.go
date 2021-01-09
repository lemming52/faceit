package publisher

import (
	"context"
	"encoding/json"
	"faceit/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
)

type SNSClient struct {
	client   *sns.SNS
	topicArn *string
}

func NewSNSClient() *SNSClient {
	client := &SNSClient{
		topicArn: aws.String("arn:aws:sns:eu-west-1:000000000000:messages_sns"), // very hardcoded
	}
	client.client = sns.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion("eu-west-1").
		WithEndpoint("http://localstack:4566"). // Hardcoded for simplicity in task
		WithDisableEndpointHostPrefix(true).
		WithDisableSSL(true).
		WithCredentials(credentials.NewStaticCredentials("dummy", "dummy", "dummy")),
	)))
	return client
}

func (pub *SNSClient) Publish(ctx context.Context, m *model.Message) error {
	msg, err := json.Marshal(m)
	if err != nil {
		return err
	}
	_, err = pub.client.PublishWithContext(ctx, &sns.PublishInput{
		Message:  aws.String(string(msg)),
		TopicArn: pub.topicArn,
	})
	if err != nil {
		return err
	}
	return nil
}
