package service

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-kit/kit/log/level"
	uuid "github.com/satori/go.uuid"
)

// AWS SDK send message to SQS
func SendSQSMessage(awsService *awsService, event cloudevents.Event) error {

	if event.Type() == "" {
		return errors.New("missing event type")
	}

	if event.ID() == "" {
		event.SetID(uuid.NewV4().String())
	}

	if event.Time().String() == "" {
		event.SetTime(time.Now())
	}

	if event.Source() == "" {
		event.SetSource("lamassu/aws-connector/" + awsService.ID)
	}

	if event.SpecVersion() == "" {
		event.SetSpecVersion("1.0")
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		level.Error(awsService.logger).Log("err", err, "msg", "Error while serializing event")
		return err
	}

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	result, err := GetSQSQueueURL(sess, awsService.SqsQueueName)
	if err != nil {
		level.Error(awsService.logger).Log("err", err, "msg", "Got an error getting the queue URL")
		return err
	}

	// Getting SQS Queue URL
	queueURL := result.QueueUrl

	level.Info(awsService.logger).Log("err", err, "msg", "Sending message type="+event.Type()+" id="+event.ID()+" queue="+*queueURL)

	sqsClient := sqs.New(sess)
	_, err = sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    queueURL,
		MessageBody: aws.String(string(eventBytes)),
		// MessageAttributes: map[string]*sqs.MessageAttributeValue{
		// 	attr: &sqs.MessageAttributeValue{
		// 		DataType:    aws.String("String"),
		// 		StringValue: aws.String(attr),
		// 	}},
	})

	if err != nil {
		level.Info(awsService.logger).Log("err", err, "msg", "Error while sending message type="+event.Type()+" id="+event.ID())
		return err
	}
	return err
}

// AWS SDK get SQS Queue URL
func GetSQSQueueURL(sess *session.Session, queue string) (*sqs.GetQueueUrlOutput, error) {
	sqsClient := sqs.New(sess)

	result, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queue,
	})
	if err != nil {
		return nil, err
	}

	return result, nil
}
