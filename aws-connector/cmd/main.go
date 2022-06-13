package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	"github.com/lamassuiot/aws-connector/pkg/server/api/transport"
	"github.com/lamassuiot/aws-connector/pkg/server/config"
	"github.com/lamassuiot/aws-connector/pkg/server/store/db"
	"github.com/opentracing/opentracing-go"

	"github.com/aws/aws-sdk-go/aws/credentials"
	awsIot "github.com/aws/aws-sdk-go/service/iot"
	awsSts "github.com/aws/aws-sdk-go/service/sts"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/lamassuiot/aws-connector/pkg/server/discovery/consul"
	lamassucaclient "github.com/lamassuiot/lamassuiot/pkg/ca/client"
	clientUtils "github.com/lamassuiot/lamassuiot/pkg/utils/client"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
)

func main() {

	// Define logger
	var logger log.Logger
	{
		logger = log.NewJSONLogger(os.Stdout)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = level.NewFilter(logger, level.AllowInfo())
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Read environment variables
	err, cfg := config.NewConfig("")
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not read environment configuration values")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Environment configuration values loaded")

	jcfg, err := jaegercfg.FromEnv()
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not load Jaeger configuration values fron environment")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Jaeger configuration values loaded")

	tracer, closer, err := jcfg.NewTracer(
		jaegercfg.Logger(jaegerlog.StdLogger),
	)
	opentracing.SetGlobalTracer(tracer)

	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not start Jaeger tracer")
		os.Exit(1)
	}
	defer closer.Close()
	level.Info(logger).Log("msg", "Jaeger tracer started")

	consulsd, err := consul.NewServiceDiscovery(cfg.ConsulProtocol, cfg.ConsulHost, cfg.ConsulPort, cfg.ConsulCA, logger)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not start connection with Consul Service Discovery")
		os.Exit(1)
	}

	level.Info(logger).Log("msg", "Connection established with Consul Service Discovery")
	connectorID, err := consulsd.Register(cfg.ConnectorProtocol, cfg.ConnectorPort, []string{cfg.ConnectorType}, cfg.ConnectorName, cfg.ConnectorPersistenceDir)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not register service liveness information to Consul")
		os.Exit(1)
	}
	level.Info(logger).Log("msg", "Service liveness information registered to Consul")

	lamassuCAClient, err := lamassucaclient.NewLamassuCAClient(clientUtils.ClientConfiguration{
		URL: &url.URL{
			Scheme: "https",
			Host:   cfg.LamassuCAAddress,
		},
		AuthMethod: clientUtils.MutualTLS,
		AuthMethodConfig: &clientUtils.MutualTLSConfig{
			ClientCert: cfg.LamassuCAClientCertFile,
			ClientKey:  cfg.LamassuCAClientKeyFile,
		},
		CACertificate: cfg.LamassuCACertFile,
	})
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create Lamassu CA Client")
		os.Exit(1)
	}

	dbStore, err := db.NewInMemoryDB()
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create InMemory DB")
		os.Exit(1)
	}

	sess := session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AwsDefaultRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AwsAccessKeyID, cfg.AwsSecretAccessKey, ""),
	}))
	awsSts := awsSts.New(sess, aws.NewConfig())
	awsSvc := awsIot.New(sess, aws.NewConfig())

	awsIdentity, err := awsSts.GetCallerIdentity(nil)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not get AWS Identity")
		os.Exit(1)
	}

	svc := service.NewAwsConnectorService(logger, connectorID, lamassuCAClient, dbStore, awsSvc, *awsIdentity.Account)
	transport.MakeSqsHandler(svc, logger, tracer, cfg.AwsDefaultRegion, *awsIdentity.Account, cfg.AwsSqsInboundQueueName)

	http.Handle("/v1/", http.StripPrefix("/v1", transport.MakeHTTPHandler(svc, log.With(logger, "component", "HTTP"), tracer)))

	http := &http.Server{
		Addr: ":" + cfg.ConnectorPort,
	}

	errs := make(chan error)
	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	http.ListenAndServe()

	level.Info(logger).Log("exit", <-errs)

}
