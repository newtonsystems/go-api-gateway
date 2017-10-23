package addsvc

// This file provides extra server-side bindings for the HTTP transport
// to all internal services

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	httptransport "github.com/go-kit/kit/transport/http"
	stdopentracing "github.com/opentracing/opentracing-go"
)

var main_logger log.Logger

// MakeHTTPHandler returns a handler that makes a set of endpoints available on predefined paths.
func MakeDebugHTTPHandler(endpoints Endpoints, tracer stdopentracing.Tracer, logger log.Logger) http.Handler {
	options := []httptransport.ServerOption{
		httptransport.ServerErrorEncoder(errorEncoder),
		httptransport.ServerErrorLogger(logger),
	}

	main_logger = logger

	m := http.NewServeMux()
	m.Handle("/sayhello", httptransport.NewServer(
		endpoints.SayHelloEndpoint,
		DecodeHTTPSayHelloRequest,
		EncodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(httptransport.PopulateRequestContext), httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "SayHello", logger)))...,
	))
	m.Handle("/getavailableagents", httptransport.NewServer(
		endpoints.GetAvailableAgentsEndpoint,
		DecodeHTTPGetAvailableAgentsRequest,
		EncodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(httptransport.PopulateRequestContext), httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "GetAvailableAgents", logger)))...,
	))
	m.Handle("/getagentidfromref", httptransport.NewServer(
		endpoints.GetAgentIDFromRefEndpoint,
		DecodeHTTPGetAgentIDFromRefRequest,
		EncodeHTTPGenericResponse,
		append(options, httptransport.ServerBefore(httptransport.PopulateRequestContext), httptransport.ServerBefore(opentracing.FromHTTPRequest(tracer, "GetAgentIDFromRef", logger)))...,
	))

	return m
}

// -- SayHello

func DecodeHTTPSayHelloRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	main_logger.Log(getRequestInfoArgs(r)...)
	var req sayHelloRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// -- GetAvailableAgents
func DecodeHTTPGetAvailableAgentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	main_logger.Log(getRequestInfoArgs(r)...)
	var req getAvailableAgentsRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// -- GetAgentIDFromRef
func DecodeHTTPGetAgentIDFromRefRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	main_logger.Log(getRequestInfoArgs(r)...)
	var req getAgentIDFromRefRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}
