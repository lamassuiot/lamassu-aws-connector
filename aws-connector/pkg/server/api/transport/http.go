package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/transport"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/lamassuiot/aws-connector/pkg/server/api/endpoint"
	"github.com/lamassuiot/aws-connector/pkg/server/api/errors"
	lamassuErrors "github.com/lamassuiot/aws-connector/pkg/server/api/errors"
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"

	stdopentracing "github.com/opentracing/opentracing-go"
)

type errorer interface {
	error() error
}

func ErrMissingDeviceID() error {
	return &lamassuErrors.GenericError{
		Message:    "Device ID not specified",
		StatusCode: 400,
	}
}

func MakeHTTPHandler(s service.Service, logger log.Logger, otTracer stdopentracing.Tracer) http.Handler {
	r := mux.NewRouter()
	e := endpoint.MakeServerEndpoints(s, otTracer)
	options := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		httptransport.ServerErrorEncoder(encodeError),
	}

	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		e.HealthEndpoint,
		decodeEmptyRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	r.Methods("GET").Path("/config").Handler(httptransport.NewServer(
		e.GetConfigurationEndpoint,
		decodeEmptyRequest,
		enocdeGetConnectorsResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Config", logger)))...,
	))

	r.Methods("GET").Path("/things/{deviceId}/config").Handler(httptransport.NewServer(
		e.GetThingsConfigurationEndpoint,
		decodeGetThingConfigRequest,
		enocdeGetThingConnectorsResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetThingsConfig", logger)))...,
	))

	r.Methods("POST").Path("/create-ca").Handler(httptransport.NewServer(
		e.SignCertificateEndpoint,
		decodeCreateCaRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "CreateCA", logger)))...,
	))

	r.Methods("PUT").Path("/ca/status").Handler(httptransport.NewServer(
		e.UpdateCAStatusEndpoint,
		decodeUpdateCaStatusRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "UpdateCAStatus", logger)))...,
	))

	r.Methods("PUT").Path("/cert/status").Handler(httptransport.NewServer(
		e.UpdateCertStatusEndpoint,
		decodeUpdateCertStatusRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "UpdateCertStatus", logger)))...,
	))
	r.Methods("POST").Path("/attach-policy").Handler(httptransport.NewServer(
		e.AttachIoTCorePolicyEndpoint,
		decodeAttachPolicyRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "AttachPolicy", logger)))...,
	))

	return r
}

func decodeEmptyRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoint.EmptyRequest
	return req, nil
}

func decodeCreateCaRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {

	var createCaRequest endpoint.RegistrationCodeRequestRequest
	json.NewDecoder(r.Body).Decode(&createCaRequest)
	if err != nil {
		return nil, err
	}

	return createCaRequest, nil
}
func decodeGetThingConfigRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	deviceId, ok := vars["deviceId"]
	if !ok {
		return nil, ErrMissingDeviceID()
	}
	var requestInput endpoint.GetThingConfigurationRequest
	requestInput.DeviceID = deviceId

	return requestInput, nil
}
func decodeUpdateCaStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {

	var deleteCaRequest endpoint.UpdateCaStatusRequest
	json.NewDecoder(r.Body).Decode(&deleteCaRequest)
	if err != nil {
		return nil, err
	}

	return deleteCaRequest, nil
}
func decodeUpdateCertStatusRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {

	var updateCertStatusRequest endpoint.UpdateCertStatusRequest
	json.NewDecoder(r.Body).Decode(&updateCertStatusRequest)
	if err != nil {
		return nil, err
	}
	return updateCertStatusRequest, nil
}
func decodeAttachPolicyRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var bindCAAwsPolicyRequest endpoint.AttachIoTCorePolicyRequest
	json.NewDecoder(r.Body).Decode(&bindCAAwsPolicyRequest)
	if err != nil {
		return nil, err
	}

	return bindCAAwsPolicyRequest, nil
}

func enocdeGetConnectorsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)

		return nil
	}
	getConfigResponse := response.(endpoint.GetConfigurationResponse)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(getConfigResponse.Config)
}

func enocdeGetThingConnectorsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)

		return nil
	}
	getConfigResponse := response.(endpoint.GetThingConfigurationResponse)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(getConfigResponse.ThingConfig)
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)

		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	// http.Error(w, err.Error(), codeFrom(err))
	w.WriteHeader(codeFrom(err))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})

}

type errorWrapper struct {
	Error string `json:"error"`
}

func codeFrom(err error) int {
	switch e := err.(type) {
	case *errors.ValidationError:
		return http.StatusBadRequest
	case *errors.DuplicateResourceError:
		return http.StatusConflict
	case *errors.ResourceNotFoundError:
		return http.StatusNotFound
	case *errors.GenericError:
		return e.StatusCode
	default:
		return http.StatusInternalServerError
	}
}
