package server

import (
	"clubmanager/api/grpc/proto"
	"clubmanager/logging"

	userService "clubmanager/services/users"

	"net"

	"github.com/jackc/pgx/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
  healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func MakeServerAndRun(addr string, db *pgx.Conn) error {
  usvc := logging.NewUserLoggingService(userService.NewUserService(db))

  clubManagerServer := NewClubManagerServiceServer(usvc)

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
  usvc userService.UserService
  proto.UnimplementedClubManagerServiceServer
}

func NewClubManagerServiceServer(usvc userService.UserService) *ClubManagerServiceServer {
  return &ClubManagerServiceServer{
    usvc: usvc,
  }
}

