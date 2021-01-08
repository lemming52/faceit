package componenttests

import (
	"encoding/json"
	"faceit/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
)

type SQSClient struct {
	Client *sqs.SQS
	Url    *string
}

func NewSQSClient(t *testing.T) *SQSClient {
	svc := sqs.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion("eu-west-1").
		WithEndpoint("http://localhost:4566"). // Hardcoded for simplicity in task
		WithDisableEndpointHostPrefix(true).
		WithDisableSSL(true).
		WithCredentials(credentials.NewStaticCredentials("dummy", "dummy", "dummy")),
	)))
	result, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String("user_messages"),
	})
	// Purge the Queue to ensure messages from other tests do not corrupt
	_, err = svc.PurgeQueue(&sqs.PurgeQueueInput{
		QueueUrl: result.QueueUrl,
	})
	assert.Nil(t, err)
	return &SQSClient{
		Client: svc,
		Url:    result.QueueUrl,
	}
}

func (s *SQSClient) deleteMessage(t *testing.T, receiptHandle *string) {
	_, err := s.Client.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      s.Url,
		ReceiptHandle: receiptHandle,
	})
	assert.Nil(t, err)
}

func TestSNSEmitted(t *testing.T) {
	payload := `{
		"forename": "Jacky",
		"surname": "Yip",
		"nickname": "Stewie2K",
		"password": "MIBR",
		"email": "jy@notarealemail.com",
		"country": "USA"
	}`
	expectedMessages := 1
	expectedAction := model.UserAdd
	svc := NewSQSClient(t)

	// Send Message by adding user
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results := &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)
	id := results.Id

	result, err := svc.Client.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:              svc.Url,
		AttributeNames:        []*string{aws.String("All")},
		MessageAttributeNames: []*string{aws.String("All")},
	})
	assert.Equal(t, expectedMessages, len(result.Messages))
	defer svc.deleteMessage(t, result.Messages[0].ReceiptHandle)

	message := &model.Message{}
	err = json.Unmarshal([]byte(*result.Messages[0].Body), message)
	assert.Nil(t, err)
	assert.Equal(t, expectedAction, message.Action)
	assert.Equal(t, id, message.Id)

	// Cleanup
	deleteCode := 204
	uri = fmt.Sprintf("%s/users/%s", getHost(), id)
	req, err = http.NewRequest(http.MethodDelete, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making delete request")
	assert.Equal(t, deleteCode, res.StatusCode)
}
