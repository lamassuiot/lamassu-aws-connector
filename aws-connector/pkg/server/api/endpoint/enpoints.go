package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint                        endpoint.Endpoint
	DispatchAttachIoTCorePolicyEndpoint   endpoint.Endpoint
	DispatchRegistrationCodeEndpoint      endpoint.Endpoint
	SignCertificateEndpoint               endpoint.Endpoint
	GetThingsConfigurationEndpoint        endpoint.Endpoint
	UpdateThingsConfigurationEndpoint     endpoint.Endpoint
	GetConfigurationEndpoint              endpoint.Endpoint
	UpdateConfigurationEndpoint           endpoint.Endpoint
	DispatchDeleteCACodeEndpoint          endpoint.Endpoint
	DispatchDeleteCertCodeEndpoint        endpoint.Endpoint
	HandleUpdateCertificateStatusEndpoint endpoint.Endpoint
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
	var dispatchDeleteCACodeEndpoint endpoint.Endpoint
	{
		dispatchDeleteCACodeEndpoint = MakeDispatchDeleteCACodeEndpoint(s)
	}
	var dispatchDeleteCertCodeEndpoint endpoint.Endpoint
	{
		dispatchDeleteCertCodeEndpoint = MakeDispatchDeleteCertCodeEndpoint(s)
	}
	var signCertificateEndpoint endpoint.Endpoint
	{
		signCertificateEndpoint = MakeSignCertificateEndpoint(s)
	}

	var getConfigurationEndpoint endpoint.Endpoint
	{
		getConfigurationEndpoint = MakeGetConfigurationEndpoint(s)
	}
	var getThingsConfigurationEndpoint endpoint.Endpoint
	{
		getThingsConfigurationEndpoint = MakeGetThingsConfigurationEndpoint(s)
	}
	var updateThingsConfigurationEndpoint endpoint.Endpoint
	{
		updateThingsConfigurationEndpoint = MakeUpdateThingsConfigurationEndpoint(s)
	}
	var updateConfigurationEndpoint endpoint.Endpoint
	{
		updateConfigurationEndpoint = MakeUpdateConfigurationEndpoint(s)
	}
	var updateCertificateStatus endpoint.Endpoint
	{
		updateCertificateStatus = MakeHandleUpdateCertificateStatusEndpoint(s)
	}
	return Endpoints{
		HealthEndpoint:                        healthEndpoint,
		DispatchAttachIoTCorePolicyEndpoint:   dispatchAttachIoTCorePolicyEndpoint,
		DispatchRegistrationCodeEndpoint:      dispatchRegistrationCodeEndpoint,
		SignCertificateEndpoint:               signCertificateEndpoint,
		GetThingsConfigurationEndpoint:        getThingsConfigurationEndpoint,
		UpdateThingsConfigurationEndpoint:     updateThingsConfigurationEndpoint,
		GetConfigurationEndpoint:              getConfigurationEndpoint,
		UpdateConfigurationEndpoint:           updateConfigurationEndpoint,
		DispatchDeleteCACodeEndpoint:          dispatchDeleteCACodeEndpoint,
		DispatchDeleteCertCodeEndpoint:        dispatchDeleteCertCodeEndpoint,
		HandleUpdateCertificateStatusEndpoint: updateCertificateStatus,
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
func MakeDispatchDeleteCACodeEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchUpdateCaStatusCodeRequest)
		err := s.DispatchUpdateCAStatusRequest(ctx, req.CaName, "")
		return "OK", err
	}
}
func MakeDispatchDeleteCertCodeEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchUpdateCertStatusCodeRequest)
		err := s.DispatchUpdateCertStatusRequest(ctx, req.DeviceID, req.SerialNumber, req.Status, req.DeviceCert, req.CaCert)
		return "OK", err
	}
}
func MakeSignCertificateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(SignRegistrationCodeRequest)
		err := s.HandleSignRegistrationCode(ctx, req.RegistrationCode, req.CaName, req.CaCert, req.SerialNumber)
		return CreateCAResponse{}, err
	}
}

func MakeGetThingsConfigurationEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetThingConfigurationRequest)
		config, err := s.GetThingConfiguration(ctx, req.DeviceID)
		return GetThingConfigurationResponse{ThingConfig: config}, err
	}
}

func MakeUpdateThingsConfigurationEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateThingConfigurationRequest)
		err := s.HandleUpdateThingConfiguration(ctx, req.DeviceID, req.Config)
		return nil, err
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
		err := s.HandleUpdateConfiguration(ctx, req.Config)
		return nil, err
	}
}
func MakeHandleUpdateCertificateStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HandleUpdateCertStatusCodeRequest)
		err := s.HandleUpdateCertificateStatus(ctx, req.CaName, req.SerialNumber, req.Status)
		return nil, err
	}
}
