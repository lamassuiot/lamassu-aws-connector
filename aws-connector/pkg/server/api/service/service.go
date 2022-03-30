package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/lamassuiot/aws-connector/pkg/server/store"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/client"
	"github.com/opentracing/opentracing-go"
)

type Service interface {
	Health(ctx context.Context) bool

	//Requests sent to AWS via SQS
	DispatchAttachIoTCorePolicy(ctx context.Context, CaName string, SerialNumber string, Policy string) error
	DispatchRegistrationCodeRequest(ctx context.Context, CaCert string, CaName string, SerialNumber string) error
	DispatchGetConfiguration(ctx context.Context) error

	//Responses received from AWS via SQS
	SignRegistrationCode(ctx context.Context, RegistrationCode string, CaName string, CaCert string, SerialNumber string) error
	UpdateConfiguration(ctx context.Context, config interface{}) error

	//
	GetConfiguration(ctx context.Context) (map[string]interface{}, error)
}

type awsService struct {
	ID              string
	logger          log.Logger
	LamassuCaClient lamassucaclient.LamassuCaClient
	SqsQueueName    string
	db              store.DB
}

func NewAwsConnectorService(logger log.Logger, connectorId string, lamassuCAClient lamassucaclient.LamassuCaClient, sqsQueueName string, db store.DB) (s Service) {
	return &awsService{
		logger:          logger,
		LamassuCaClient: lamassuCAClient,
		SqsQueueName:    sqsQueueName,
		ID:              connectorId,
		db:              db,
	}
}

func (s *awsService) Health(ctx context.Context) bool {
	return true
}

func (s *awsService) DispatchAttachIoTCorePolicy(ctx context.Context, CaName string, SerialNumber string, Policy string) error {

	attachPolicyContent := struct {
		CaName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
		Policy       string `json:"policy"`
	}{
		CaName:       CaName,
		SerialNumber: SerialNumber,
		Policy:       Policy,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.policy.attach")
	event.SetData(cloudevents.ApplicationJSON, attachPolicyContent)

	// Sending message to SQS
	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return err
}

func (s *awsService) DispatchRegistrationCodeRequest(ctx context.Context, CaCert string, CaName string, SerialNumber string) error {
	initializeCARegistration := struct {
		CaCert       string `json:"ca_cert"`
		CaName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
	}{
		CaName:       CaName,
		CaCert:       CaCert,
		SerialNumber: SerialNumber,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.registration.init")
	event.SetData(cloudevents.ApplicationJSON, initializeCARegistration)

	// Sending message to SQS
	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}
	return nil
}

func (s *awsService) DispatchGetConfiguration(ctx context.Context) error {
	emptyData := struct{}{}

	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.config.request")
	event.SetData(cloudevents.ApplicationJSON, emptyData)

	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}

	return nil
}

func (s *awsService) GetConfiguration(ctx context.Context) (map[string]interface{}, error) {
	getConfigurationWithRetry := func(ctx context.Context) map[string]interface{} {
		for {
			config, err := s.db.GetAWSIoTCoreConfig(ctx)
			fmt.Print(err)
			if err == nil {
				// s.db.DeleteAWSIoTCoreConfig(ctx)
				return config
			}

			time.Sleep(1 * time.Second)
		}
	}

	if _, err := s.db.GetAWSIoTCoreConfig(ctx); err != nil {
		s.DispatchGetConfiguration(ctx)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	config := make(chan map[string]interface{}, 1)
	go func() {
		config <- getConfigurationWithRetry(ctx)
	}()
	select {
	case conf := <-config:
		return conf, nil
	case <-ctx.Done():
		return make(map[string]interface{}), errors.New("timeout while geting config")
	}
}

func (s *awsService) UpdateConfiguration(ctx context.Context, config interface{}) error {
	s.db.UpdateAWSIoTCoreConfig(ctx, config)
	return nil
}

func (s *awsService) SignRegistrationCode(ctx context.Context, RegistrationCode string, CaName string, CaCert string, SerialNumber string) error {
	// Generate CSR template for CA verification certificate
	subj := pkix.Name{
		CommonName: RegistrationCode,
	}
	rawSubj := subj.ToRDNSequence()
	asn1Subj, _ := asn1.Marshal(rawSubj)
	template := x509.CertificateRequest{
		RawSubject:         asn1Subj,
		SignatureAlgorithm: x509.SHA512WithRSA,
	}

	// Generate Private key for verification certificate
	privKey, _ := rsa.GenerateKey(rand.Reader, 4096)
	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privKey)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to create registraion code CSR")
		return err
	}

	// Generate CSR for verification certificate
	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to parse registraion code CSR")
		return err
	}

	ctx = context.Background()
	ctx = context.WithValue(ctx, "LamassuLogger", s.logger)
	sapn, ctx := opentracing.StartSpanFromContext(ctx, "SignCertificateRequest")
	fmt.Println(sapn)

	// Sign verification certificate CSR
	crt, err := s.LamassuCaClient.SignCertificateRequest(ctx, CaName, csr, "pki", false)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to sign registration code.")
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})

	finializeCARegistrationProcessContent := struct {
		CaCert           string `json:"ca_cert"`
		CaName           string `json:"ca_name"`
		SerialNumber     string `json:"serial_number"`
		VerificationCert string `json:"verification_cert"`
	}{
		CaName:           CaName,
		SerialNumber:     SerialNumber,
		CaCert:           CaCert,
		VerificationCert: base64.StdEncoding.EncodeToString(certPEM),
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.registration.signed_challenge")
	event.SetData(cloudevents.ApplicationJSON, finializeCARegistrationProcessContent)

	// Sending message to SQS
	err = SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}

	s.db.DeleteAWSIoTCoreConfig(ctx)

	return nil
}
