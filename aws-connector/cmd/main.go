package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	"github.com/lamassuiot/aws-connector/pkg/server/api/transport"
	"github.com/lamassuiot/aws-connector/pkg/server/config"
	"github.com/lamassuiot/aws-connector/pkg/server/store/db"
	"github.com/opentracing/opentracing-go"

	"github.com/lamassuiot/aws-connector/pkg/server/discovery/consul"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	lamassucaclient "github.com/lamassuiot/lamassu-ca/client"
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

	lamassuCAClient, err := lamassucaclient.NewLamassuCaClient(cfg.LamassuCAAddress, cfg.LamassuCACertFile, cfg.LamassuCAClientCertFile, cfg.LamassuCAClientKeyFile, logger)
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create Lamassu CA Client")
		os.Exit(1)
	}

	dbStore, err := db.NewInMemoryDB()
	if err != nil {
		level.Error(logger).Log("err", err, "msg", "Could not create InMemory DB")
		os.Exit(1)
	}

	svc := service.NewAwsConnectorService(logger, connectorID, lamassuCAClient, cfg.AwsSqsQueueName, dbStore)
	transport.MakeSqsHandler(svc, logger, tracer)

	//mux := http.NewServeMux()

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
