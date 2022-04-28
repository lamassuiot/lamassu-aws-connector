module github.com/lamassuiot/aws-connector

go 1.16

// replace github.com/lamassuiot/lamassu-ca => /home/ikerlan/lamassu/lamassu-ca

require (
	github.com/aws/aws-sdk-go v1.42.35
	github.com/cloudevents/sdk-go/v2 v2.6.0
	github.com/dgraph-io/badger/v3 v3.2103.2
	github.com/go-kit/kit v0.12.0
	github.com/go-kit/log v0.2.0
	github.com/gorilla/mux v1.8.0
	github.com/hashicorp/consul/api v1.10.1
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/lamassuiot/lamassu-ca v1.0.21
	github.com/opentracing/opentracing-go v1.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/uber/jaeger-client-go v2.25.0+incompatible
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/sys v0.0.0-20220209214540-3681064d5158 // indirect
)
