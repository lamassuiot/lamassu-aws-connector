package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/lamassuiot/aws-connector/pkg/server/api/endpoint"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	log "github.com/sirupsen/logrus"
)

var (
	sqsSvc *sqs.SQS
)

func pollMessages(chn chan<- *sqs.Message, sqsURL string) {

	log.Info("Polling messages from SQS")
	for {
		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsURL),
			MaxNumberOfMessages: aws.Int64(2),
			WaitTimeSeconds:     aws.Int64(15),
		})

		if err != nil {
			log.Warn("Error polling messages from SQS: ", err)
		}

		for _, message := range output.Messages {
			chn <- message
		}
	}

}

func handleMessage(msg *sqs.Message, e endpoint.Endpoints) error {
	var event cloudevents.Event
	json.Unmarshal([]byte(*msg.Body), &event)

	jsonM, _ := event.MarshalJSON()
	log.Debug(fmt.Sprintf("Incoming SQS message. SQS Message ID %s. Event type %s. Event JSON: %s", *msg.MessageId, event.Type(), string(jsonM)))

	switch event.Type() {
	case "io.lamassu.iotcore.cert.status.update":
		var eventData endpoint.HandleUpdateCertStatusCodeRequest
		json.Unmarshal(event.Data(), &eventData)
		log.Info(eventData)
		_, err := e.HandleUpdateCertificateStatusEndpoint(context.Background(), eventData)
		return err

	case "io.lamassu.iotcore.ca.status.update":
		var eventData endpoint.HandleUpdateCAStatusCodeRequest
		json.Unmarshal(event.Data(), &eventData)
		_, err := e.HandleUpdateCAStatusEndpoint(context.Background(), eventData)
		log.Info(eventData)
		return err

	default:
		log.Warn(fmt.Sprintf("no matching evene type for incoming SQS message with type:%s ", event.Type()))
		return errors.New("unhandeled message")
	}
}

func deleteMessage(msg *sqs.Message, sqsURL string) {
	sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(sqsURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	log.Trace(fmt.Sprintf("Deleted message from SQS with Message ID: %s", *msg.MessageId))
}

func MakeSQSHandler(s service.Service, awsSQSInboundQueueName string) {
	e := endpoint.MakeServerEndpoints(s)
	awsRegion := s.GetDefaultRegion()
	awsAccountID := s.GetAccountID()
	sqsURL := "https://sqs." + awsRegion + ".amazonaws.com/" + awsAccountID + "/" + awsSQSInboundQueueName

	go func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String(awsRegion)}}))

		sqsSvc = sqs.New(sess)

		chnMessages := make(chan *sqs.Message, 2)
		go pollMessages(chnMessages, sqsURL)

		for message := range chnMessages {
			err := handleMessage(message, e)
			if err == nil {
				deleteMessage(message, sqsURL)
			}
		}
	}()
}
