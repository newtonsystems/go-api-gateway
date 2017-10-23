package addsvc

// This file contains methods to make individual endpoints from services,
// request and response types to serve those endpoints, as well as encoders and
// decoders for those types, for all of our supported transport serialization
// formats. It also includes endpoint middlewares.

import (
	"context"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/newtonsystems/grpc_types/go/grpc_types"
	"google.golang.org/grpc"
	"time"
)

// Endpoints collects all of the endpoints that compose an add service. It's
// meant to be used as a helper struct, to collect all of the endpoints into a
// single parameter.
//
// In a server, it's useful for functions that need to operate on a per-endpoint
// basis. For example, you might pass an Endpoints to a function that produces
// an http.Handler, with each method (endpoint) wired up to a specific path. (It
// is probably a mistake in design to invoke the Service methods on the
// Endpoints struct in a server.)
//
// In a client, it's useful to collect individually constructed endpoints into a
// single type that implements the Service interface. For example, you might
// construct individual endpoints using transport/http.NewClient, combine them
// into an Endpoints, and return it to the caller as a Service.

// Endpoint is the fundamental building block of servers and clients.
// It represents a single RPC method.

type Endpoints struct {
	SayHelloEndpoint           endpoint.Endpoint
	SayWorldEndpoint           endpoint.Endpoint
	GetAvailableAgentsEndpoint endpoint.Endpoint
	GetAgentIDFromRefEndpoint  endpoint.Endpoint
}

// TODO: It is super inefficient to connect via newclient everytime
// needs sorting to be faster
func MakeSayHelloEndpoint(connection *grpc.ClientConn) endpoint.Endpoint {
	client := grpc_types.NewHelloClient(connection)

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		sayHelloReq := request.(sayHelloRequest)

		resp, err := client.SayHello(
			ctx,
			&grpc_types.HelloRequest{Name: sayHelloReq.Name},
		)
		var msg string = ""
		if resp != nil {
			msg = resp.Message
		}

		return sayHelloResponse{Message: msg, Err: err}, err
	}
}

func MakeGetAvailableAgentsEndpoint(connection *grpc.ClientConn) endpoint.Endpoint {
	client := grpc_types.NewAgentMgmtClient(connection)

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		getAvailableAgentsReq := request.(getAvailableAgentsRequest)
		resp, err := client.GetAvailableAgents(
			ctx,
			&grpc_types.GetAvailableAgentsRequest{Limit: getAvailableAgentsReq.Limit},
		)

		var agentIDs []string
		if resp != nil {
			agentIDs = resp.AgentIds
		}

		return getAvailableAgentsResponse{AgentIds: agentIDs, Err: err}, err
	}
}

func MakeGetAgentIDFromRefEndpoint(connection *grpc.ClientConn) endpoint.Endpoint {
	client := grpc_types.NewAgentMgmtClient(connection)

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		getAgentIDFromRefReq := request.(getAgentIDFromRefRequest)
		resp, err := client.GetAgentIDFromRef(
			ctx,
			&grpc_types.GetAgentIDFromRefRequest{RefId: getAgentIDFromRefReq.RefId},
		)

		var agentID int32
		if resp != nil {
			agentID = resp.AgentId
		}

		return getAgentIDFromRefResponse{AgentId: agentID, Err: err}, err
	}
}

// -- MIDDLEWARE / LOGGING STUFF --

// EndpointInstrumentingMiddleware returns an endpoint middleware that records
// the duration of each invocation to the passed histogram. The middleware adds
// a single field: "success", which is "true" if no error is returned, and
// "false" otherwise.
func EndpointInstrumentingMiddleware(duration metrics.Histogram) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			defer func(begin time.Time) {
				duration.With("success", fmt.Sprint(err == nil)).Observe(time.Since(begin).Seconds())
			}(time.Now())
			return next(ctx, request)

		}
	}
}

// EndpointLoggingMiddleware returns an endpoint middleware that logs the
// duration of each invocation, and the resulting error, if any.
func EndpointLoggingMiddleware(logger log.Logger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				logger.Log("level", "err", "error", err, "took", time.Since(begin))
			}(time.Now())
			return next(ctx, request)

		}
	}
}

// These types are unexported because they only exist to serve the endpoint
// domain, which is totally encapsulated in this package. They are otherwise
// opaque to all callers.

// NOTE: json package only stringifies fields start with capital letter.
// see http://golang.org/pkg/encoding/json/
// i.e. Uses first letter caps for fields

type sayHelloRequest struct{ Name string }

type sayHelloResponse struct {
	Message string
	Err     error
}

type sayWorldRequest struct{ Name string }

type sayWorldResponse struct {
	Message string
}

// GetAvailableAgents()
type getAvailableAgentsRequest struct {
	Limit int32
}

type getAvailableAgentsResponse struct {
	AgentIds []string
	Err      error
}

// GetAgentIDFromRef()
type getAgentIDFromRefRequest struct {
	RefId string
}

type getAgentIDFromRefResponse struct {
	AgentId int32
	Err     error
}
