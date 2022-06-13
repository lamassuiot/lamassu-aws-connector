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
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/lamassuiot/aws-connector/pkg/server/api/endpoint"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	stdopentracing "github.com/opentracing/opentracing-go"
)

var (
	sqsSvc *sqs.SQS
)

func pollMessages(chn chan<- *sqs.Message, logger log.Logger, sqsURL string) {

	level.Info(logger).Log("msg", "Polling messages from SQS")

	for {
		output, err := sqsSvc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(sqsURL),
			MaxNumberOfMessages: aws.Int64(2),
			WaitTimeSeconds:     aws.Int64(15),
		})

		if err != nil {
			level.Error(logger).Log("err", err, "msg", "Error polling messages from SQS")
		}

		for _, message := range output.Messages {
			chn <- message
		}
	}

}

func handleMessage(msg *sqs.Message, e endpoint.Endpoints, logger log.Logger) error {
	var event cloudevents.Event
	json.Unmarshal([]byte(*msg.Body), &event)

	level.Info(logger).Log("msg", "Received message from SQS type="+event.Type())
	jsonM, _ := event.MarshalJSON()
	fmt.Println(string(jsonM))

	switch event.Type() {
	case "io.lamassu.iotcore.cert.status.update":
		var eventData endpoint.HandleUpdateCertStatusCodeRequest
		json.Unmarshal(event.Data(), &eventData)
		level.Info(logger).Log("msg", eventData)
		_, err := e.HandleUpdateCertificateStatusEndpoint(context.Background(), eventData)
		level.Debug(logger).Log("msg", eventData)
		return err

	case "io.lamassu.iotcore.ca.status.update":
		var eventData endpoint.HandleUpdateCAStatusCodeRequest
		json.Unmarshal(event.Data(), &eventData)
		_, err := e.HandleUpdateCAStatusEndpoint(context.Background(), eventData)
		level.Debug(logger).Log("msg", eventData)
		return err

	default:
		level.Error(logger).Log("msg", "no matching evene type for incoming SQS message with type: "+event.Type())
		return errors.New("unhandeled message")
	}

	return nil
}

func deleteMessage(msg *sqs.Message, logger log.Logger, sqsURL string) {
	sqsSvc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      aws.String(sqsURL),
		ReceiptHandle: msg.ReceiptHandle,
	})
	level.Info(logger).Log("msg", "Deleted message from SQS")
}

func MakeSqsHandler(s service.Service, logger log.Logger, otTracer stdopentracing.Tracer, awsRegion string, awsAccountID, awsSQSInboundQueueName string) {
	e := endpoint.MakeServerEndpoints(s, otTracer)
	sqsURL := "https://sqs." + awsRegion + ".amazonaws.com/" + awsAccountID + "/" + awsSQSInboundQueueName
	go func() {
		sess := session.Must(session.NewSessionWithOptions(session.Options{Config: aws.Config{Region: aws.String("eu-west-1")}}))

		sqsSvc = sqs.New(sess)

		chnMessages := make(chan *sqs.Message, 2)
		go pollMessages(chnMessages, logger, sqsURL)

		for message := range chnMessages {
			err := handleMessage(message, e, logger)
			if err == nil {
				deleteMessage(message, logger, sqsURL)
			}
		}
	}()
}
