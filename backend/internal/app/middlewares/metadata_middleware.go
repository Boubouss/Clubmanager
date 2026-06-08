package middlewares

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func MetadataInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if vals := md.Get("client-ip"); len(vals) > 0 {
			ctx = context.WithValue(ctx, "client_ip", vals[0])
		}
		if vals := md.Get("user-agent"); len(vals) > 0 {
			ctx = context.WithValue(ctx, "user_agent", vals[0])
		}
	}

	return handler(ctx, req)
}
