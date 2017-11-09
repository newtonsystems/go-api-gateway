package addsvc

// This file provides server-side bindings for the gRPC transport.
// It utilizes the transport/grpc.Server.

import (
	"context"

	"github.com/go-kit/kit/log"
	oldcontext "golang.org/x/net/context"
	//"github.com/go-kit/kit/tracing/opentracing"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"

	"github.com/newtonsystems/grpc_types/go/grpc_types"
)

func MakeAllServicesGRPCServer(endpoints Endpoints, tracer stdopentracing.Tracer, logger log.Logger) grpc_types.GlobalAPIServer {
	//options := []grpctransport.ServerOption{
	//	grpctransport.ServerErrorLogger(logger),
	//}
	return &grpcAllServicesServer{
		sayhello: grpctransport.NewServer(
			endpoints.SayHelloEndpoint,
			DecodeGRPCSayHelloRequest,
			EncodeGRPCSayHelloResponse,
			//append(options, grpctransport.ServerBefore(opentracing.FromGRPCRequest(tracer, "Sum", logger)))...,
		),
		sayworld: grpctransport.NewServer(
			endpoints.SayWorldEndpoint,
			DecodeGRPCSayHelloRequest,
			EncodeGRPCSayHelloResponse,
			//append(options, grpctransport.ServerBefore(opentracing.FromGRPCRequest(tracer, "Sum", logger)))...,
		),
	}
}

type grpcAllServicesServer struct {
	sayhello           grpctransport.Handler
	sayworld           grpctransport.Handler
	getavailableagents grpctransport.Handler
	getagentidfromref  grpctransport.Handler
	acceptcall         grpctransport.Handler
	heartbeat          grpctransport.Handler
	addtask            grpctransport.Handler
	ping               grpctransport.Handler
}

// -- Hello Service
func (s *grpcAllServicesServer) Ping(ctx oldcontext.Context, req *grpc_types.PingRequest) (*grpc_types.PingResponse, error) {
	_, rep, err := s.getavailableagents.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.PingResponse), nil
}

func (s *grpcAllServicesServer) GetAvailableAgents(ctx oldcontext.Context, req *grpc_types.GetAvailableAgentsRequest) (*grpc_types.GetAvailableAgentsResponse, error) {
	_, rep, err := s.getavailableagents.ServeGRPC(ctx, req)

	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.GetAvailableAgentsResponse), nil
}

func (s *grpcAllServicesServer) GetAgentIDFromRef(ctx oldcontext.Context, req *grpc_types.GetAgentIDFromRefRequest) (*grpc_types.GetAgentIDFromRefResponse, error) {
	_, rep, err := s.getagentidfromref.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.GetAgentIDFromRefResponse), nil
}

func (s *grpcAllServicesServer) SayHello(ctx oldcontext.Context, req *grpc_types.HelloRequest) (*grpc_types.HelloResponse, error) {
	_, rep, err := s.sayhello.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.HelloResponse), nil
}

func (s *grpcAllServicesServer) SayWorld(ctx oldcontext.Context, req *grpc_types.WorldRequest) (*grpc_types.WorldResponse, error) {
	_, rep, err := s.sayworld.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.WorldResponse), nil
}

func (s *grpcAllServicesServer) AcceptCall(ctx oldcontext.Context, req *grpc_types.AcceptCallRequest) (*grpc_types.AcceptCallResponse, error) {
	_, rep, err := s.acceptcall.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.AcceptCallResponse), nil
}

func (s *grpcAllServicesServer) HeartBeat(ctx oldcontext.Context, req *grpc_types.HeartBeatRequest) (*grpc_types.HeartBeatResponse, error) {
	_, rep, err := s.heartbeat.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.HeartBeatResponse), nil
}

func (s *grpcAllServicesServer) AddTask(ctx oldcontext.Context, req *grpc_types.AddTaskRequest) (*grpc_types.AddTaskResponse, error) {
	_, rep, err := s.addtask.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*grpc_types.AddTaskResponse), nil
}

// Decode SayHello response i.e from hello service to go-kit structure endpoint
func DecodeGRPCSayHelloResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*grpc_types.HelloResponse)
	return sayHelloResponse{Message: reply.Message}, nil
}

// go-kit -> hello request service
// Encode from go-kit request to hello service message

// -- Hello Service

func DecodeGRPCSayHelloRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*grpc_types.HelloRequest)
	return sayHelloRequest{Name: req.Name}, nil
}

func EncodeGRPCSayHelloResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(sayHelloResponse)
	return &grpc_types.HelloResponse{Message: resp.Message}, nil
}
