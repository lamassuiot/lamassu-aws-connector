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
	"github.com/lamassuiot/aws-connector/pkg/server/api/service"

	stdopentracing "github.com/opentracing/opentracing-go"
)

type errorer interface {
	error() error
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
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	r.Methods("POST").Path("/create-ca").Handler(httptransport.NewServer(
		e.DispatchRegistrationCodeEndpoint,
		decodeCreateCaRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	r.Methods("POST").Path("/attach-policy").Handler(httptransport.NewServer(
		e.DispatchAttachIoTCorePolicyEndpoint,
		decodeAttachPolicyRequest,
		encodeResponse,
		append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "Health", logger)))...,
	))

	return r
}

func decodeEmptyRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var req endpoint.EmptyRequest
	return req, nil
}

func decodeCreateCaRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {

	var createCaRequest endpoint.DispatchRegistrationCodeRequestRequest
	json.NewDecoder(r.Body).Decode(&createCaRequest)
	if err != nil {
		return nil, err
	}

	return createCaRequest, nil
}

func decodeAttachPolicyRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	var bindCAAwsPolicyRequest endpoint.DispatchAttachIoTCorePolicyRequest
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
