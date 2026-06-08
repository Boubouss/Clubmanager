package server

import (
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// toGRPCStatus maps a Go error to an appropriate gRPC status error.
// Internal errors are not exposed to the caller.
func toGRPCStatus(err error) error {
	if err == nil {
		return nil
	}
	// Pass through already-wrapped gRPC status errors (except Unknown which we remap).
	if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
		return err
	}

	msg := err.Error()
	switch {
	case strings.Contains(msg, "unauthorized"):
		return status.Error(codes.PermissionDenied, msg)
	case strings.Contains(msg, "not found"):
		return status.Error(codes.NotFound, msg)
	case strings.Contains(msg, "already exists"), strings.Contains(msg, "already registered"), strings.Contains(msg, "already a passenger"):
		return status.Error(codes.AlreadyExists, msg)
	case strings.Contains(msg, "invalid"), strings.Contains(msg, "must be"), strings.Contains(msg, "required"):
		return status.Error(codes.InvalidArgument, msg)
	case strings.Contains(msg, "no seats"), strings.Contains(msg, "period is closed"), strings.Contains(msg, "not open"):
		return status.Error(codes.FailedPrecondition, msg)
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
