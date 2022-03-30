package config

import "github.com/kelseyhightower/envconfig"

// RabbitMQ
// Esto podria ser la config de un consumer
// Puede que cambie para publisher
type Config struct {
	Host string `required:"true" split_words:"true"`
	Port string `required:"true" split_words:"true"`

	ConnectorPort           string `required:"true" split_words:"true"`
	ConnectorType           string `required:"true" split_words:"true"`
	ConnectorProtocol       string `required:"true" split_words:"true"`
	ConnectorName           string `required:"true" split_words:"true"`
	ConnectorPersistenceDir string `required:"true" split_words:"true"`

	User           string
	Password       string
	Exchange       string
	Queue          string
	RoutingKey     string
	ConsumerTag    string
	WorkerPoolSize int

	ConsulProtocol string `required:"true" split_words:"true"`
	ConsulHost     string `required:"true" split_words:"true"`
	ConsulPort     string `required:"true" split_words:"true"`
	ConsulCA       string `required:"true" split_words:"true"`

	AwsSqsQueueName string `required:"true" split_words:"true"`

	LamassuCACertFile       string `required:"true" split_words:"true"`
	LamassuCAClientCertFile string `required:"true" split_words:"true"`
	LamassuCAClientKeyFile  string `required:"true" split_words:"true"`
	LamassuCAAddress        string `required:"true" split_words:"true"`
}

func NewConfig(prefix string) (error, Config) {
	var cfg Config
	err := envconfig.Process(prefix, &cfg)
	if err != nil {
		return err, Config{}
	}
	return nil, cfg
}
