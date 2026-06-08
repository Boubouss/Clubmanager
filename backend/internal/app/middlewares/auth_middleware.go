package middlewares

import (
	"clubmanager/internal/app/services"
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var publicRoutes = map[string]bool{
	"/proto.ClubManagerService/CreateUser": true,
	"/proto.ClubManagerService/LoginUser":  true,
}

func AuthInterceptor(tkm services.TokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicRoutes[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		vals := md.Get("authorization")
		if len(vals) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization token")
		}

		tokenStr := strings.TrimPrefix(vals[0], "Bearer ")
		userId, err := tkm.ParseToken(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "user_id", userId)
		return handler(ctx, req)
	}
}
