export AWS_ACCESS_KEY_ID=foobar
export AWS_SECRET_ACCESS_KEY=foobar

aws dynamodb create-table \
--endpoint-url=http://localhost:4566 \
--region eu-west-1 \
--table-name faceit-users \
--attribute-definitions AttributeName=userId,AttributeType=S \
--key-schema AttributeName=userId,KeyType=HASH \
--provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5

aws dynamodb batch-write-item \
--endpoint-url=http://localhost:4566 \
--region eu-west-1 \
--request-items file://testdata/users.json

aws sns create-topic --name messages_sns --endpoint-url=http://localhost:4566

aws sqs create-queue --endpoint-url=http://localhost:4566 --queue-name user_messages

aws --endpoint-url=http://localhost:4566 sns subscribe \
--topic-arn arn:aws:sns:eu-west-1:000000000000:messages_sns \
--protocol sqs \
--notification-endpoint http://localhost:4566/queue/user_messages \
--attributes RawMessageDelivery=true
