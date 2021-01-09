package dao

import (
	"context"
	"errors"
	"faceit/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
)

type DynamoClient struct {
	client       *dynamodb.DynamoDB
	table        *string
	partitionKey string
	decoder      *dynamodbattribute.Decoder
	encoder      *dynamodbattribute.Encoder
}

func NewDynamoClient() *DynamoClient {
	client := &DynamoClient{
		table:        aws.String("faceit-users"),
		partitionKey: "userId",
	}
	client.decoder = dynamodbattribute.NewDecoder()
	client.encoder = dynamodbattribute.NewEncoder()

	client.client = dynamodb.New(session.Must(session.NewSession(aws.NewConfig().
		WithRegion("eu-west-1").
		WithEndpoint("http://localstack:4566"). // Hardcoded for simplicity in task
		WithDisableEndpointHostPrefix(true).
		WithDisableSSL(true).
		WithCredentials(credentials.NewStaticCredentials("dummy", "dummy", "dummy")),
	)))
	return client
}

func (db *DynamoClient) Get(ctx context.Context, id string) (*model.User, error) {
	user := &model.User{}
	input := dynamodb.GetItemInput{
		TableName: db.table,
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {S: aws.String(id)},
		},
	}

	res, err := db.client.GetItemWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}
	if len(res.Item) == 0 {
		return nil, errors.New("no such user")
	}
	db.decode(res.Item, user)
	return user, nil
}

func (db *DynamoClient) Insert(ctx context.Context, user *model.User) error {
	attr, err := db.encode(user)
	if err != nil {
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: db.table,
		Item:      attr,
	}
	_, err = db.client.PutItemWithContext(ctx, input)
	return err
}

func (db *DynamoClient) Delete(ctx context.Context, id string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: db.table,
		Key: map[string]*dynamodb.AttributeValue{
			db.partitionKey: {S: aws.String(id)},
		},
	}
	_, err := db.client.DeleteItemWithContext(ctx, input)
	return err
}

// small convenience function
func (db *DynamoClient) decode(output map[string]*dynamodb.AttributeValue, object interface{}) {
	attr := &dynamodb.AttributeValue{
		M: output,
	}
	db.decoder.Decode(attr, object)
}

func (db *DynamoClient) encode(object interface{}) (map[string]*dynamodb.AttributeValue, error) {
	attr, err := db.encoder.Encode(object)
	if err != nil {
		return nil, err
	}
	return attr.M, nil
}

func (db *DynamoClient) Filter(ctx context.Context, conditions []*model.FilterCondition) ([]*model.User, error) {
	var filters []expression.ConditionBuilder
	for _, condition := range conditions {
		filters = append(filters, expression.Name(condition.Query).Equal(expression.Value(condition.Value)))
	}
	filt := combineFilters(filters)
	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}
	input := &dynamodb.ScanInput{
		TableName:                 db.table,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		FilterExpression:          expr.Filter(),
	}
	res, err := db.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, nil
	}
	users := []*model.User{}
	for _, item := range res.Items {
		user := &model.User{}
		db.decode(item, user)
		users = append(users, user)
	}
	return users, nil
}

func combineFilters(filters []expression.ConditionBuilder) expression.ConditionBuilder {
	switch len(filters) {
	case 1:
		return filters[0]
	case 2:
		return expression.And(filters[0], filters[1])
	default:
		return expression.And(filters[0], filters[1], filters[2:]...)
	}
}

func (db *DynamoClient) GetAll(ctx context.Context) ([]*model.User, error) {
	input := &dynamodb.ScanInput{
		TableName: db.table,
	}
	res, err := db.client.ScanWithContext(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(res.Items) == 0 {
		return nil, nil
	}
	users := []*model.User{}
	for _, item := range res.Items {
		user := &model.User{}
		db.decode(item, user)
		users = append(users, user)
	}
	return users, nil
}
