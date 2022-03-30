package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/log/level"
)

type AwsConnectorClient interface {
	RegisterCA(ctx context.Context, caName string, caSerialNumber string, caCertificate string) error
	AttachAccessPolicy(ctx context.Context, caName string, caSerialNumber string, serializedAccessPolicy string) error
	GetConfiguration(ctx context.Context) (AWSConfig, error)
}
type AwsConnectorClientConfig struct {
	client BaseClient
	logger log.Logger
	ID     string
}

func NewAwsConnectorClient(id string, ip string, port string, logger log.Logger) (AwsConnectorClient, error) {
	client := &http.Client{}
	url := ip + ":" + port

	return &AwsConnectorClientConfig{
		logger: logger,
		client: NewBaseClient(url, client),
		ID:     id,
	}, nil
}

func (s *AwsConnectorClientConfig) RegisterCA(ctx context.Context, caName string, caSerialNumber string, caCertificate string) error {
	level.Info(s.logger).Log("msg", "Resgitering CA to AWS")

	awsCreateIotCoreCA := awsCreateIotCoreCA{
		CaName:       caName,
		CaCert:       caCertificate,
		SerialNumber: caSerialNumber,
	}
	awsCreateIotCoreCABytes, err := json.Marshal(awsCreateIotCoreCA)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}

	req, err := s.client.NewRequest("POST", "/v1/create-ca", awsCreateIotCoreCABytes)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}
	_, _, err = s.client.Do(req)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}

	return nil
}

func (s *AwsConnectorClientConfig) AttachAccessPolicy(ctx context.Context, caName string, caSerialNumber string, serializedAccessPolicy string) error {
	fmt.Println("Calling attach access policy", s.ID, caName, serializedAccessPolicy)

	awsIotCoreCAAttachPolicy := awsIotCoreCAAttachPolicy{
		Policy:       serializedAccessPolicy,
		CaName:       caName,
		SerialNumber: caSerialNumber,
	}
	fmt.Println(awsIotCoreCAAttachPolicy)

	awsCreateIotCoreCABytes, err := json.Marshal(awsIotCoreCAAttachPolicy)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}

	req, err := s.client.NewRequest("POST", "/v1/attach-policy", awsCreateIotCoreCABytes)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}
	_, _, err = s.client.Do(req)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return err
	}

	return nil
}

func (s *AwsConnectorClientConfig) GetConfiguration(ctx context.Context) (AWSConfig, error) {
	var config AWSConfig
	req, err := s.client.NewRequest("GET", "/v1/config", nil)

	if err != nil {
		level.Error(s.logger).Log("err", err)
		return AWSConfig{}, err
	}

	config, _, err = s.client.Do(req)
	if err != nil {
		level.Error(s.logger).Log("err", err)
		return AWSConfig{}, err
	}

	return config, nil
}
