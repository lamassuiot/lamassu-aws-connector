# AWS Connector

[< :house:](.)

### Table of contents
* [Overview](#Overview)
* [Prerequisites](#Prerequisites)
* [Usage](#Usage)
* [References](#References)
* [Authors](#Authors)
* [License](#License)  

## Overview

This repo gathers all the code that has been developed to integrate Lamassu IOT and Amazon Web Services.

## Prerequisites

1. ***Option 1***: Install go. [(Go installation guide)](https://go.dev/doc/install)
2. ***Option 2***: run the AWS connector in Docker. [(Docker installation guide)](https://docs.docker.com/engine/install/)

## Usage

Needed environment variables:

```env
# AWS credentials
AWS_ACCESS_KEY_ID=
AWS_SECRET_ACCESS_KEY=
AWS_DEFAULT_REGION=

# AWS ATS root certificate
AWS_CA_BUNDLE=awsRootCA.pem

# RabbitMQ configuration: port, host and certificates
AMQP_PORT=5671
AMQP_HOST=dev-lamassu.zpd.ikerlan.es
AMQP_SERVER_CA_CERT=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/ca.crt
AMQP_CLIENT_CERT=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/aws-connector/tls.crt
AMQP_CLIENT_KEY=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/aws-connector/tls.key

# Lamassu CA configuration
LAMASSU_CA_ADDRESS=https://lamassu-ca:8087
LAMASSU_CA_CERT_FILE=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/ca.crt
LAMASSU_CA_CLIENT_CERT_FILE=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/aws-connector/tls.crt
LAMASSU_CA_CLIENT_KEY_FILE=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/aws-connector/tls.key

# Consul configuration
CONSUL_PROTOCOL=https
CONSUL_HOST=consul-server
CONSUL_PORT=8501
CONSUL_CA=/home/ikerlan/lamassu-compose-v2/tls-certificates/upstream/ca.crt

# AWS Connector configuration
CONNECTOR_PORT=8989
CONNECTOR_TYPE=aws
CONNECTOR_PROTOCOL=http
CONNECTOR_NAME=aws-connector-ikerlan

# Configuration for jaeger tracing
JAEGER_SERVICE_NAME=aws-connector
JAEGER_AGENT_HOST=jaeger
JAEGER_AGENT_PORT="6831"
JAEGER_SAMPLER_TYPE=const
JAEGER_SAMPLER_PARAM="1"
JAEGER_REPORTER_LOG_SPANS="false"
```
> :warning: ***Remember to change env variables depending where you run AWS connector***

***Option 1***: Running AWS Connector

```bash
cd aws-connector
go run cmd/main.go
```

***Option 2***: Running the connector on Docker:

```bash
cd aws-connector
docker-compose up -d
```

> ****NOTE****: It is better to run it on Docker.

## References

* Gokit, building microservices in go: [(Gokit)](https://gokit.io/faq/)

## Authors

- Haritz Saiz @hsaiz – [hsaiz@ikerlan.es](mailto:hsaiz@ikerlan.es)
- Markel Orallo @morallo – [morallo@ikerlan.es](mailto:morallo@ikerlan.es)

## License

[< :house:](.)