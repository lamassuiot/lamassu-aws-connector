package endpoint

import (
	"context"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/go-kit/kit/endpoint"
	api "github.com/lamassuiot/aws-connector/pkg/common"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"
	cProviderEndpoint "github.com/lamassuiot/lamassuiot/pkg/cloud-provider/server/api/endpoint"
)

type Endpoints struct {
	cProviderEndpoint.Endpoints
	HandleUpdateCertificateStatusEndpoint endpoint.Endpoint
	HandleUpdateCAStatusEndpoint          endpoint.Endpoint
	HandleCloudEvents                     endpoint.Endpoint
}

func MakeServerEndpoints(s service.Service) Endpoints {
	e := cProviderEndpoint.MakeServerEndpoints(s)

	updateCAtatus := MakeHandleUpdateCAStatusEndpoint(s)
	updateCertificateStatus := MakeHandleUpdateCertificateStatusEndpoint(s)
	cloudEvents := MakeHandleCloudEvents(s)

	return Endpoints{
		Endpoints: cProviderEndpoint.Endpoints{
			RegisterCAEndpoint:                                e.RegisterCAEndpoint,
			UpdateConfigurationEndpoint:                       e.UpdateConfigurationEndpoint,
			GetConfigurationEndpoint:                          e.GetConfigurationEndpoint,
			GetDeviceConfigurationEndpoint:                    e.GetDeviceConfigurationEndpoint,
			UpdateCAStatusEndpoint:                            e.UpdateCAStatusEndpoint,
			UpdateDeviceCertificateStatusEndpoint:             e.UpdateDeviceCertificateStatusEndpoint,
			UpdateDeviceDigitalTwinReenrollmentStatusEndpoint: e.UpdateDeviceDigitalTwinReenrollmentStatusEndpoint,
		},
		HandleUpdateCertificateStatusEndpoint: updateCertificateStatus,
		HandleUpdateCAStatusEndpoint:          updateCAtatus,
		HandleCloudEvents:                     cloudEvents,
	}
}

func MakeHandleUpdateCAStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HandleUpdateCAStatusCodeRequest)
		err := s.HandleUpdateCAStatus(ctx, &api.HandleUpdateCAStatusInput{
			CaName:         req.CaName,
			CaSerialNumber: req.CaSerialNumber,
			CaID:           req.CaID,
			Status:         req.Status,
		})
		return nil, err
	}
}

func MakeHandleCloudEvents(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		event := request.(cloudevents.Event)
		err = s.HandleCloudEvents(ctx, event)
		return nil, err
	}
}
func MakeHandleUpdateCertificateStatusEndpoint(s service.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(HandleUpdateCertStatusCodeRequest)
		err := s.HandleUpdateCertificateStatus(ctx, &api.HandleUpdateCertificateStatusInput{
			CaName:       req.CaName,
			SerialNumber: req.SerialNumber,
			Status:       req.Status,
		})
		return nil, err
	}
}
