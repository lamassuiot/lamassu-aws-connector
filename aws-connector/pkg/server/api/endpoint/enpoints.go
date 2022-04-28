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
	DispatchUpdateCAStatusEndpoint        endpoint.Endpoint
	DispatchUpdateCertStatusEndpoint      endpoint.Endpoint
	HandleUpdateCertificateStatusEndpoint endpoint.Endpoint
	HandleUpdateCAStatusEndpoint          endpoint.Endpoint
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
	var dispatchUpdateCAStatusEndpoint endpoint.Endpoint
	{
		dispatchUpdateCAStatusEndpoint = MakeDispatchUpdateCAStatusEndpoint(s)
	}
	var dispatchUpdateCertStatusEndpoint endpoint.Endpoint
	{
		dispatchUpdateCertStatusEndpoint = MakeDispatchUpdateCertStatusEndpoint(s)
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
	var updateCAtatus endpoint.Endpoint
	{
		updateCAtatus = MakeHandleUpdateCAStatusEndpoint(s)
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
		DispatchUpdateCAStatusEndpoint:        dispatchUpdateCAStatusEndpoint,
		DispatchUpdateCertStatusEndpoint:      dispatchUpdateCertStatusEndpoint,
		HandleUpdateCertificateStatusEndpoint: updateCertificateStatus,
		HandleUpdateCAStatusEndpoint:          updateCAtatus,
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
func MakeDispatchUpdateCAStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchUpdateCaStatusRequest)
		err := s.DispatchUpdateCAStatusRequest(ctx, req.CaName, req.Status, req.CertificateID)
		return "OK", err
	}
}
func MakeDispatchUpdateCertStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(DispatchUpdateCertStatusRequest)
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

func MakeHandleUpdateCAStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HandleUpdateCAStatusCodeRequest)
		err := s.HandleUpdateCAStatus(ctx, req.CaName, req.CaSerialNumber, req.CaID, req.Status)
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
