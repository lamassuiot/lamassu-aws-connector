package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	"github.com/lamassuiot/aws-connector/pkg/server/api/transport"
	"github.com/lamassuiot/aws-connector/pkg/server/config"
	"github.com/lamassuiot/aws-connector/pkg/server/store/db"
	"github.com/lamassuiot/lamassuiot/pkg/cloud-provider/server/api/discovery"
	cloudprovidertransport "github.com/lamassuiot/lamassuiot/pkg/cloud-provider/server/api/transport"
	lamassudevmanager "github.com/lamassuiot/lamassuiot/pkg/device-manager/client"
	lamassudmsclient "github.com/lamassuiot/lamassuiot/pkg/dms-manager/client"
	clientUtils "github.com/lamassuiot/lamassuiot/pkg/utils/client"
	"github.com/lamassuiot/lamassuiot/pkg/utils/server"
	"github.com/lamassuiot/lamassuiot/pkg/v3/clients"
	configV3 "github.com/lamassuiot/lamassuiot/pkg/v3/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	config := config.NewAWSConnectorConfig()
	mainServer := server.NewServer(config)

	var clientConf configV3.HTTPClient
	if strings.HasPrefix(config.LamassuCAAddress, "https") {
		clientConf = configV3.HTTPClient{
			AuthMode: configV3.MTLS,
			AuthMTLSOptions: configV3.AuthMTLSOptions{
				CertFile: config.CertFile,
				KeyFile:  config.KeyFile,
			},
			HTTPConnection: configV3.HTTPConnection{
				Protocol: configV3.HTTPS,
				BasePath: "",
				BasicConnection: configV3.BasicConnection{
					TLSConfig: configV3.TLSConfig{
						InsecureSkipVerify: true,
						CACertificateFile:  config.LamassuCACertFile,
					},
				},
			},
		}
	} else {
		clientConf = configV3.HTTPClient{
			AuthMode: configV3.NoAuth,
			HTTPConnection: configV3.HTTPConnection{
				Protocol: configV3.HTTP,
				BasePath: "",
				BasicConnection: configV3.BasicConnection{
					TLSConfig: configV3.TLSConfig{
						InsecureSkipVerify: true,
						CACertificateFile:  config.LamassuCACertFile,
					},
				},
			},
		}
	}

	caHttpClient, err := clients.BuildHTTPClient(clientConf, "CA")
	if err != nil {
		log.Fatal(err)
	}

	caClient := clients.NewhttpCAClient(caHttpClient, config.LamassuCAAddress)

	var dmsClient lamassudmsclient.LamassuDMSManagerClient
	parsedLamassuDMSURL, err := url.Parse(config.LamassuDMSAddress)
	if err != nil {
		log.Fatal("Could not parse LamassuDMS url: ", err)
	}

	if strings.HasPrefix(config.LamassuDMSAddress, "https") {
		dmsClient, err = lamassudmsclient.NewLamassuDMSManagerClientConfig(clientUtils.BaseClientConfigurationuration{
			URL:        parsedLamassuDMSURL,
			AuthMethod: clientUtils.AuthMethodMutualTLS,
			AuthMethodConfig: &clientUtils.MutualTLSConfig{
				ClientCert: config.CertFile,
				ClientKey:  config.KeyFile,
			},
			CACertificate: config.LamassuDMSCertFile,
			Insecure:      config.LamassuDMSInsecureSkipVerify,
		})
		if err != nil {
			log.Fatal("Could not create LamassuCA client: ", err)
		}
	} else {
		dmsClient, err = lamassudmsclient.NewLamassuDMSManagerClientConfig(clientUtils.BaseClientConfigurationuration{
			URL:        parsedLamassuDMSURL,
			AuthMethod: clientUtils.AuthMethodNone,
		})
		if err != nil {
			log.Fatal("Could not create LamassuCA client: ", err)
		}
	}

	var devManagerClient lamassudevmanager.LamassuDeviceManagerClient
	parsedLamassuDevManagerURL, err := url.Parse(config.LamassuDeviceManagerAddress)
	if err != nil {
		log.Fatal("Could not parse Device Manager URL: ", err)
	}

	if strings.HasPrefix(config.LamassuDeviceManagerAddress, "https") {
		devManagerClient, err = lamassudevmanager.NewLamassuDeviceManagerClient(clientUtils.BaseClientConfigurationuration{
			URL:        parsedLamassuDevManagerURL,
			AuthMethod: clientUtils.AuthMethodMutualTLS,
			AuthMethodConfig: &clientUtils.MutualTLSConfig{
				ClientCert: config.CertFile,
				ClientKey:  config.KeyFile,
			},
			CACertificate: config.LamassuDeviceManagerCertFile,
		})
		if err != nil {
			log.Fatal("Could not create Device Manager client: ", err)
		}
	} else {
		devManagerClient, err = lamassudevmanager.NewLamassuDeviceManagerClient(clientUtils.BaseClientConfigurationuration{
			URL:        parsedLamassuDevManagerURL,
			AuthMethod: clientUtils.AuthMethodNone,
		})
		if err != nil {
			log.Fatal("Could not create Device Manager client: ", err)
		}
	}

	consulsd, err := discovery.NewServiceDiscovery(config.ConsulProtocol, config.ConsulHost, config.ConsulPort, config.ConsulCA, config.ConsulInsecureSkipVerify)
	if err != nil {
		log.Fatal("Could not create Consul client: ", err)
	}

	healthy, _, err := consulsd.CheckHealth()
	if err != nil {
		log.Fatal("Could not check Consul health: ", err)
	}

	if !healthy {
		log.Fatal("Consul is not healthy")
	} else {
		log.Info("Consul is healthy")
	}

	connectorID, err := consulsd.Register(config.Protocol, config.Port, []string{config.ConnectorType}, config.ConnectorName, config.ConnectorPersistenceDir)
	if err != nil {
		log.Fatal("Could not register service liveness information to Consul: ", err)
	}
	log.Info(fmt.Sprintf("Service liveness information registered to Consul. ID: %s", connectorID))

	dbStore, err := db.NewInMemoryDB()
	if err != nil {
		log.Fatal("Could not create InMemory DB: ", err)
	}

	svc, err := service.NewAwsConnectorService(connectorID, caClient, dmsClient, devManagerClient, dbStore, config.AWSDefaultRegion, config.AWSAccessKeyID, config.AWSSecretAccessKey, config.AWSSqsOutboundQueueName)
	if err != nil {
		log.Fatal("Could not create AWS Connector Service: ", err)
	}

	svc = service.LoggingMiddleware()(svc)

	mainServer.AddHttpHandler("/v1/", http.StripPrefix("/v1", cloudprovidertransport.MakeHTTPHandler(svc)))
	// transport.MakeSQSHandler(svc, config.AWSSqsInboundQueueName)
	mainServer.AddAmqpConsumer(config.ServiceName, []string{"#"}, transport.MakeAmqpHandler(svc))

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	mainServer.Run()
	forever := make(chan struct{})
	<-forever
}
