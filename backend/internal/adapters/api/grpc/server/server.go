package server

import (
	"clubmanager/internal/adapters/api/grpc/proto"
	"clubmanager/internal/app/middlewares"
	"clubmanager/internal/app/services"
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

type ClubManagerServices struct {
	UserService    services.UserService
	ClubService    services.ClubService
	MemberService  services.MemberService
	LicenceService services.LicenceService
	RoleService    services.RoleService
	EventService   services.EventService
}

type ClubManagerServiceServer struct {
	usvc services.UserService
	csvc services.ClubService
	msvc services.MemberService
	lsvc services.LicenceService
	rsvc services.RoleService
	esvc services.EventService
	proto.UnimplementedClubManagerServiceServer
}

func NewClubManagerServiceServer(svc *ClubManagerServices) *ClubManagerServiceServer {
	return &ClubManagerServiceServer{
		usvc: svc.UserService,
		csvc: svc.ClubService,
		msvc: svc.MemberService,
		lsvc: svc.LicenceService,
		rsvc: svc.RoleService,
		esvc: svc.EventService,
	}
}

// errorMappingInterceptor converts Go errors to appropriate gRPC status codes.
func errorMappingInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	resp, err := handler(ctx, req)
	if err != nil {
		return nil, toGRPCStatus(err)
	}
	return resp, nil
}

// recoveryInterceptor catches panics and converts them to Internal errors.
func recoveryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}

// MakeServerAndRun starts the gRPC server and blocks until ctx is cancelled (graceful shutdown).
func MakeServerAndRun(ctx context.Context, addr string, svc *ClubManagerServices, tkm services.TokenManager) error {
	clubManagerServer := NewClubManagerServiceServer(svc)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			middlewares.MetadataInterceptor,
			middlewares.AuthInterceptor(tkm),
			errorMappingInterceptor,
			recoveryInterceptor,
		),
	)

	proto.RegisterClubManagerServiceServer(grpcServer, clubManagerServer)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("proto.ClubManagerService", healthpb.HealthCheckResponse_SERVING)

	errCh := make(chan error, 1)
	go func() {
		errCh <- grpcServer.Serve(ln)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		healthServer.SetServingStatus("proto.ClubManagerService", healthpb.HealthCheckResponse_NOT_SERVING)
		grpcServer.GracefulStop()
		return nil
	}
}
