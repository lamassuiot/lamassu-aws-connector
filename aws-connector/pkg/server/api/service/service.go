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
	"fmt"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	lamassuErrors "github.com/lamassuiot/aws-connector/pkg/server/api/errors"
	"github.com/lamassuiot/aws-connector/pkg/server/store"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/pkg/client"

	"github.com/opentracing/opentracing-go"
)

type Service interface {
	Health(ctx context.Context) bool

	//Requests sent to AWS via SQS
	DispatchAttachIoTCorePolicy(ctx context.Context, caName string, SerialNumber string, Policy string) error
	DispatchRegistrationCodeRequest(ctx context.Context, CaCert string, caName string, SerialNumber string) error
	DispatchGetConfiguration(ctx context.Context) error
	DispatchGetThingConfiguration(ctx context.Context, deviceID string) error
	DispatchUpdateCAStatusRequest(ctx context.Context, caName string, status string, certificateID string) error
	DispatchUpdateCertStatusRequest(ctx context.Context, caName string, certSerialNumber string, status string, deviceCert string, caCert string) error

	//Responses received from AWS via SQS
	HandleSignRegistrationCode(ctx context.Context, RegistrationCode string, caName string, CaCert string, SerialNumber string) error
	HandleUpdateConfiguration(ctx context.Context, config interface{}) error
	HandleUpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error
	HandleUpdateCertificateStatus(ctx context.Context, caName string, serialNumber string, status string) error
	HandleUpdateCAStatus(ctx context.Context, caName string, caSerialNumber string, caID string, status string) error

	// HTTP methods
	GetConfiguration(ctx context.Context) (map[string]interface{}, error)
	GetThingConfiguration(ctx context.Context, deviceID string) (AWSThing, error)
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
		CaName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
		Policy       string `json:"policy"`
	}{
		CaName:       caName,
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
		CaName       string `json:"ca_name"`
		SerialNumber string `json:"serial_number"`
	}{
		CaName:       caName,
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

func (s *awsService) DispatchUpdateCAStatusRequest(ctx context.Context, caName string, status string, certificateID string) error {
	ctx = context.WithValue(ctx, "LamassuLogger", s.logger)
	_, ctx = opentracing.StartSpanFromContext(ctx, "SignCertificateRequest")
	if status == "REVOKED" {
		status = "INACTIVE"
	}
	updateStatus := struct {
		CertificateID string `json:"certificate_id"`
		Status        string `json:"status"`
	}{
		CertificateID: certificateID,
		Status:        status,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.ca.status.update")
	event.SetData(cloudevents.ApplicationJSON, updateStatus)

	// Sending message to SQS
	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}
	err = s.db.DeleteAWSIoTCoreConfig(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *awsService) DispatchUpdateCertStatusRequest(ctx context.Context, deviceID string, certSerialNumber string, status string, deviceCert string, caCert string) error {
	ctx = context.WithValue(ctx, "LamassuLogger", s.logger)
	_, ctx = opentracing.StartSpanFromContext(ctx, "SignCertificateRequest")
	updateStatus := CertUpdateStatus{
		SerialNumber: certSerialNumber,
		DeviceID:     deviceID,
		Status:       status,
		DeviceCert:   deviceCert,
		CaCert:       caCert,
	}
	event := cloudevents.NewEvent()
	event.SetType("io.lamassu.iotcore.cert.status.update")
	event.SetData(cloudevents.ApplicationJSON, updateStatus)

	// Sending message to SQS
	err := SendSQSMessage(s, event)
	if err != nil {
		level.Error(s.logger).Log("err", err, "msg", "Failed to send message to AWS SQS")
		return err
	}

	s.db.DeleteAWSIoTCoreThingConfig(ctx, deviceID)
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
	var stop bool
	getConfigurationWithRetry := func(ctx context.Context) map[string]interface{} {
		var config map[string]interface{}
		var err error
		for {
			if stop {
				break
			}
			config, err = s.db.GetAWSIoTCoreConfig(ctx)
			// fmt.Print(err)
			if err == nil {
				// s.db.DeleteAWSIoTCoreConfig(ctx)
				return config
			}

			time.Sleep(1 * time.Second)
		}
		return config
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
		stop = true
		return make(map[string]interface{}), &lamassuErrors.GenericError{
			StatusCode: 404,
			Message:    "timeout while geting config from AWS",
		}
	}
}

func (s *awsService) GetThingConfiguration(ctx context.Context, deviceID string) (AWSThing, error) {
	var stop bool
	getConfigurationWithRetry := func(ctx context.Context) interface{} {
		var config interface{}
		var err error
		for {
			if stop {
				break
			}
			config, err = s.db.GetAWSIoTCoreThingConfig(ctx, deviceID)
			if err == nil {
				return config
			}

			time.Sleep(1 * time.Second)
		}
		return config
	}

	if _, err := s.db.GetAWSIoTCoreThingConfig(ctx, deviceID); err != nil {
		s.DispatchGetThingConfiguration(ctx, deviceID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	config := make(chan interface{}, 1)
	go func() {
		config <- getConfigurationWithRetry(ctx)
	}()
	select {
	case conf := <-config:
		var thing AWSThing
		confBytes, err := json.Marshal(conf)
		if err != nil {
			s.db.DeleteAWSIoTCoreThingConfig(ctx, deviceID)
			return AWSThing{}, err
		}
		json.Unmarshal(confBytes, &thing)
		return thing, nil
	case <-ctx.Done():
		stop = true
		return AWSThing{}, &lamassuErrors.GenericError{
			StatusCode: 404,
			Message:    "timeout while geting device config from AWS",
		}
	}
}

func (s *awsService) HandleUpdateCAStatus(ctx context.Context, caName string, caSerialNumber string, caID string, status string) error {
	level.Info(s.logger).Log("msg", "invalidating config cache due to CA update. caName:"+caName+" caSerialNumber:"+caSerialNumber+" caID:"+caID+" status:"+status)
	s.db.DeleteAWSIoTCoreConfig(ctx)
	return nil
}

func (s *awsService) HandleUpdateConfiguration(ctx context.Context, config interface{}) error {
	s.db.UpdateAWSIoTCoreConfig(ctx, config)
	return nil
}

func (s *awsService) HandleUpdateThingConfiguration(ctx context.Context, deviceID string, config interface{}) error {
	s.db.UpdateAWSIoTCoreThingConfig(ctx, deviceID, config)
	return nil
}

func (s *awsService) HandleSignRegistrationCode(ctx context.Context, RegistrationCode string, caName string, CaCert string, SerialNumber string) error {
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
		CaName           string `json:"ca_name"`
		SerialNumber     string `json:"serial_number"`
		VerificationCert string `json:"verification_cert"`
	}{
		CaName:           caName,
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
func (s *awsService) HandleUpdateCertificateStatus(ctx context.Context, caName string, serialNumber string, status string) error {
	ctx = context.WithValue(ctx, "LamassuLogger", s.logger)
	if status == "REVOKED" {
		err := s.LamassuCaClient.RevokeCert(ctx, caName, serialNumber, "pki")
		if err != nil {
			switch err.(type) {
			case *lamassucaclient.AlreadyRevokedError:
				return nil
			default:
				return err
			}

		}
	}
	return nil
}
