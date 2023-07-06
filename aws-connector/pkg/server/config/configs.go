package config

import "github.com/lamassuiot/lamassuiot/pkg/utils/server"

type AWSConnectorConfig struct {
	server.BaseConfiguration

	ConnectorType           string `required:"true" split_words:"true"`
	ConnectorName           string `required:"true" split_words:"true"`
	ConnectorPersistenceDir string `required:"true" split_words:"true"`

	Exchange string
	Queue    string

	ConsulProtocol           string `required:"true" split_words:"true"`
	ConsulHost               string `required:"true" split_words:"true"`
	ConsulPort               string `required:"true" split_words:"true"`
	ConsulCA                 string `required:"true" split_words:"true"`
	ConsulInsecureSkipVerify bool   `required:"true" split_words:"true"`

	AWSAccessKeyID          string `required:"true" split_words:"true"`
	AWSSecretAccessKey      string `required:"true" split_words:"true"`
	AWSDefaultRegion        string `required:"true" split_words:"true"`
	AWSSqsInboundQueueName  string `required:"true" split_words:"true"`
	AWSSqsOutboundQueueName string `split_words:"true"`

	LamassuCAAddress                       string `required:"true" split_words:"true"`
	LamassuCACertFile                      string `split_words:"true"`
	LamassuCAInsecureSkipVerify            bool   `required:"true" split_words:"true"`
	LamassuDMSAddress                      string `required:"true" split_words:"true"`
	LamassuDMSCertFile                     string `split_words:"true"`
	LamassuDMSInsecureSkipVerify           bool   `required:"true" split_words:"true"`
	LamassuDeviceManagerAddress            string `required:"true" split_words:"true"`
	LamassuDeviceManagerCertFile           string `split_words:"true"`
	LamassuDeviceManagerInsecureSkipVerify bool   `required:"true" split_words:"true"`
}

func NewAWSConnectorConfig() *AWSConnectorConfig {
	return &AWSConnectorConfig{}
}

func (c *AWSConnectorConfig) GetBaseConfiguration() *server.BaseConfiguration {
	return &c.BaseConfiguration
}

func (c *AWSConnectorConfig) GetConfiguration() interface{} {
	return c
}
