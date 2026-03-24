package server

import (
	"clubmanager/internal/adapters/api/grpc/proto"
	"clubmanager/internal/app/services"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

)

type ClubManagerServices struct {
  UserService services.UserService
}

type ClubManagerServiceServer struct {
  usvc services.UserService
  proto.UnimplementedClubManagerServiceServer
}

func NewClubManagerServiceServer(svc *ClubManagerServices) *ClubManagerServiceServer {
  return &ClubManagerServiceServer{
    usvc: svc.UserService,
  }
}

func MakeServerAndRun(addr string, svc *ClubManagerServices) error {
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

