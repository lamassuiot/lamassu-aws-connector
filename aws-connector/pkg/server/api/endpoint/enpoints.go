package endpoint

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/lamassuiot/lamassu-aws-connector/aws-connector/pkg/server/api/service"
	stdopentracing "github.com/opentracing/opentracing-go"
)

type Endpoints struct {
	HealthEndpoint                        endpoint.Endpoint
	AttachIoTCorePolicyEndpoint           endpoint.Endpoint
	SignCertificateEndpoint               endpoint.Endpoint
	GetThingsConfigurationEndpoint        endpoint.Endpoint
	UpdateThingsConfigurationEndpoint     endpoint.Endpoint
	GetConfigurationEndpoint              endpoint.Endpoint
	UpdateConfigurationEndpoint           endpoint.Endpoint
	UpdateCAStatusEndpoint                endpoint.Endpoint
	UpdateCertStatusEndpoint              endpoint.Endpoint
	HandleUpdateCertificateStatusEndpoint endpoint.Endpoint
	HandleUpdateCAStatusEndpoint          endpoint.Endpoint
}

func MakeServerEndpoints(s service.Service, otTracer stdopentracing.Tracer) Endpoints {

	var healthEndpoint endpoint.Endpoint
	{
		healthEndpoint = MakeHealthEndpoint(s)
	}

	var AttachIoTCorePolicyEndpoint endpoint.Endpoint
	{
		AttachIoTCorePolicyEndpoint = MakeAttachIoTCorePolicyEndpoint(s)
	}

	var UpdateCAStatusEndpoint endpoint.Endpoint
	{
		UpdateCAStatusEndpoint = MakeUpdateCAStatusEndpoint(s)
	}
	var UpdateCertStatusEndpoint endpoint.Endpoint
	{
		UpdateCertStatusEndpoint = MakeUpdateCertStatusEndpoint(s)
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
		AttachIoTCorePolicyEndpoint:           AttachIoTCorePolicyEndpoint,
		SignCertificateEndpoint:               signCertificateEndpoint,
		GetThingsConfigurationEndpoint:        getThingsConfigurationEndpoint,
		UpdateThingsConfigurationEndpoint:     updateThingsConfigurationEndpoint,
		GetConfigurationEndpoint:              getConfigurationEndpoint,
		UpdateConfigurationEndpoint:           updateConfigurationEndpoint,
		UpdateCAStatusEndpoint:                UpdateCAStatusEndpoint,
		UpdateCertStatusEndpoint:              UpdateCertStatusEndpoint,
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

func MakeAttachIoTCorePolicyEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AttachIoTCorePolicyRequest)
		err := s.AttachIoTCorePolicy(ctx, req.CaName, req.SerialNumber, req.Policy)
		return AttachIoTCorePolicyResponse{}, err
	}
}

func MakeUpdateCAStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateCaStatusRequest)
		err := s.UpdateCAStatusRequest(ctx, req.CaName, req.Status, req.CertificateID)
		return "OK", err
	}
}
func MakeUpdateCertStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(UpdateCertStatusRequest)
		err := s.UpdateCertStatusRequest(ctx, req.DeviceID, req.SerialNumber, req.Status, req.DeviceCert, req.CaCert)
		return "OK", err
	}
}

func MakeSignCertificateEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(RegistrationCodeRequestRequest)
		err := s.SignRegistrationCode(ctx, req.CaName, req.CaCert, req.SerialNumber)
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
