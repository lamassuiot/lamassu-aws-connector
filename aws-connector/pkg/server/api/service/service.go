package service

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/json"
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
	DispatchAttachIoTCorePolicy(ctx context.Context, caName string, SerialNumber string, Policy string) error
	DispatchRegistrationCodeRequest(ctx context.Context, CaCert string, caName string, SerialNumber string) error
	DispatchGetConfiguration(ctx context.Context) error
	DispatchGetThingConfiguration(ctx context.Context, deviceID string) error
	DispatchUpdaCAStatusCodeRequest(ctx context.Context, caName string, status string) error
	DispatchUpdateCertStatusCodeRequest(ctx context.Context, caName string, certSerialNumber string, status string) error

	//Responses received from AWS via SQS
	SignRegistrationCode(ctx context.Context, RegistrationCode string, caName string, CaCert string, SerialNumber string) error
	UpdateConfiguration(ctx context.Context, config interface{}) error
	UpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error

	// HTTP methods
	GetConfiguration(ctx context.Context) (map[string]interface{}, error)
	GetThingConfiguration(ctx context.Context, deviceID string) (AWSThingConfig, error)
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

func (s *awsService) DispatchAttachIoTCorePolicy(ctx context.Context, caName string, SerialNumber string, Policy string) error {

	attachPolicyContent := struct {
		caName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
		Policy       string `json:"policy"`
	}{
		caName:       caName,
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

func (s *awsService) DispatchRegistrationCodeRequest(ctx context.Context, CaCert string, caName string, SerialNumber string) error {
	initializeCARegistration := struct {
		CaCert       string `json:"ca_cert"`
		caName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
	}{
		caName:       caName,
		CaCert:       CaCert,
		SerialNumber: SerialNumber,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.registration.request-code")
	event.SetData(cloudevents.ApplicationJSON, initializeCARegistration)

	// Sending message to SQS
	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}
	return nil
}
func (s *awsService) DispatchUpdaCAStatusCodeRequest(ctx context.Context, caName string, status string) error {
	return nil
}
func (s *awsService) DispatchUpdateCertStatusCodeRequest(ctx context.Context, caName string, certSerialNumber string, status string) error {
	ctx = context.WithValue(ctx, "LamassuLogger", s.logger)
	_, ctx = opentracing.StartSpanFromContext(ctx, "SignCertificateRequest")
	deviceCert, err := s.LamassuCaClient.GetCert(ctx, caName, certSerialNumber, "pki")
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "could not get the device certificate")
		return err
	}

	deviceID := deviceCert.Subject.CN

	thingConfig, err := s.GetThingConfiguration(ctx, deviceID)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "could not get the aws thing config")
		return err
	}

	fmt.Println(deviceID, thingConfig)
	for _, thingCert := range thingConfig.Certificates {
		if thingCert.SerialNumber == certSerialNumber {
			updateStatus := struct {
				CertificateID string `json:"certificate_id"`
				Status        string `json:"status"`
			}{
				CertificateID: thingCert.ID,
				Status:        status,
			}
			event := cloudevents.NewEvent()
			event.SetType("io.lamassu.iotcore.cert.status.update")
			event.SetData(cloudevents.ApplicationJSON, updateStatus)

			// Sending message to SQS
			err = SendSQSMessage(s, event)
			if err != nil {
				level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
				return err
			}
		}
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

func (s *awsService) DispatchGetThingConfiguration(ctx context.Context, deviceID string) error {
	getConfig := struct {
		DeviceID string `json:"device_id"`
	}{
		DeviceID: deviceID,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.thing.config.request")
	event.SetData(cloudevents.ApplicationJSON, getConfig)

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
			// fmt.Print(err)
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

func (s *awsService) GetThingConfiguration(ctx context.Context, deviceID string) (AWSThingConfig, error) {
	getConfigurationWithRetry := func(ctx context.Context) []interface{} {
		for {
			config, err := s.db.GetAWSIoTCoreThingConfig(ctx, deviceID)
			fmt.Print(err)
			if err == nil {
				s.db.DeleteAWSIoTCoreThingConfig(ctx, deviceID)
				return config
			}

			time.Sleep(1 * time.Second)
		}
	}

	s.DispatchGetThingConfiguration(ctx, deviceID)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	config := make(chan []interface{}, 1)
	go func() {
		config <- getConfigurationWithRetry(ctx)
	}()
	select {
	case conf := <-config:
		var thingConfig AWSThingConfig
		confBytes, err := json.Marshal(conf)
		if err != nil {
			s.db.DeleteAWSIoTCoreThingConfig(ctx, deviceID)
			return AWSThingConfig{}, err
		}
		json.Unmarshal(confBytes, &thingConfig)
		return thingConfig, nil
	case <-ctx.Done():
		return AWSThingConfig{}, errors.New("timeout while geting config")
	}
}

func (s *awsService) UpdateConfiguration(ctx context.Context, config interface{}) error {
	s.db.UpdateAWSIoTCoreConfig(ctx, config)
	return nil
}

func (s *awsService) UpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error {
	s.db.UpdateAWSIoTCoreThingConfig(ctx, deviceID, config)
	return nil
}

func (s *awsService) SignRegistrationCode(ctx context.Context, RegistrationCode string, caName string, CaCert string, SerialNumber string) error {
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
	crt, err := s.LamassuCaClient.SignCertificateRequest(ctx, caName, csr, "pki", false)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to sign registration code.")
		return err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: crt.Raw})

	finializeCARegistrationProcessContent := struct {
		CaCert           string `json:"ca_cert"`
		caName           string `json:"ca_name"`
		SerialNumber     string `json:"serial_number"`
		VerificationCert string `json:"verification_cert"`
	}{
		caName:           caName,
		SerialNumber:     SerialNumber,
		CaCert:           CaCert,
		VerificationCert: base64.StdEncoding.EncodeToString(certPEM),
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.registration.signed-code")
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
