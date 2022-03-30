package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint                      endpoint.Endpoint
	DispatchAttachIoTCorePolicyEndpoint endpoint.Endpoint
	DispatchRegistrationCodeEndpoint    endpoint.Endpoint
	SignCertificateEndpoint             endpoint.Endpoint
	GetConfigurationEndpoint            endpoint.Endpoint
	UpdateConfigurationEndpoint         endpoint.Endpoint
}

func MakeServerEndpoints(s service.Service, otTracer stdopentracing.Tracer) Endpoints {

	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
	}

	var dispatchAttachIoTCorePolicyEndpoint endpoint.Endpoint
	{
		dispatchAttachIoTCorePolicyEndpoint = MakeDispatchAttachIoTCorePolicyEndpoint(s)
	}

	var dispatchRegistrationCodeEndpoint endpoint.Endpoint
	{
		dispatchRegistrationCodeEndpoint = MakeDispatchRegistrationCodeEndpoint(s)
	}

	var signCertificateEndpoint endpoint.Endpoint
	{
		signCertificateEndpoint = MakeSignCertificateEndpoint(s)
	}

	var getConfigurationEndpoint endpoint.Endpoint
	{
		getConfigurationEndpoint = MakeGetConfigurationEndpoint(s)
	}
	var updateConfigurationEndpoint endpoint.Endpoint
	{
		updateConfigurationEndpoint = MakeUpdateConfigurationEndpoint(s)
	}

	return Endpoints{
		HealthEndpoint:                      healthEndpoint,
		DispatchAttachIoTCorePolicyEndpoint: dispatchAttachIoTCorePolicyEndpoint,
		DispatchRegistrationCodeEndpoint:    dispatchRegistrationCodeEndpoint,
		SignCertificateEndpoint:             signCertificateEndpoint,
		GetConfigurationEndpoint:            getConfigurationEndpoint,
		UpdateConfigurationEndpoint:         updateConfigurationEndpoint,
	}
}

func MakeHealthEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		healthy := s.Health(ctx)
		return HealthResponse{Healthy: healthy}, nil
	}
}

func MakeDispatchAttachIoTCorePolicyEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchAttachIoTCorePolicyRequest)
		err := s.DispatchAttachIoTCorePolicy(ctx, req.CaName, req.SerialNumber, req.Policy)
		return DispatchAttachIoTCorePolicyResponse{}, err
	}
}

func MakeDispatchRegistrationCodeEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchRegistrationCodeRequestRequest)
		err := s.DispatchRegistrationCodeRequest(ctx, req.CaCert, req.CaName, req.SerialNumber)
		return CreateCAResponse{}, err
	}
}

func MakeSignCertificateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SignRegistrationCodeRequest)
		err := s.SignRegistrationCode(ctx, req.RegistrationCode, req.CaName, req.CaCert, req.SerialNumber)
		return CreateCAResponse{}, err
	}
}

func MakeGetConfigurationEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		config, err := s.GetConfiguration(ctx)
		return GetConfigurationResponse{Config: config}, err
	}
}

func MakeUpdateConfigurationEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateConfigurationRequest)
		err := s.UpdateConfiguration(ctx, req.Config)
		return GetConfigurationResponse{}, err
	}
}
