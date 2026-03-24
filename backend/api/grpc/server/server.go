package server

import (
	"clubmanager/api/grpc/proto"
	"clubmanager/services/users"

	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Services struct {
  UserService users.UserService
}

func MakeServerAndRun(addr string, svc *Services) error {
  clubManagerServer := NewClubManagerServiceServer(svc)

  ln, err := net.Listen("tcp", addr)
  if err != nil {
    return err
  }

  opts := []grpc.ServerOption{}
  server := grpc.NewServer(opts...)

  proto.RegisterClubManagerServiceServer(server, clubManagerServer)
  
  healthServer := health.NewServer()
  healthpb.RegisterHealthServer(server, healthServer)
  healthServer.SetServingStatus("proto.ClubManagerService", healthpb.HealthCheckResponse_SERVING)
  
  return server.Serve(ln)
}

type ClubManagerServiceServer struct {
  usvc users.UserService
  proto.UnimplementedClubManagerServiceServer
}

func NewClubManagerServiceServer(svc *Services) *ClubManagerServiceServer {
  return &ClubManagerServiceServer{
    usvc: svc.UserService,
  }
}

