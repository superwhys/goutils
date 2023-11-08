package lg

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
)

func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	td := TimeFuncDuration()

	prefix := fmt.Sprintf("[%s]", strings.TrimPrefix(info.FullMethod, "/"))

	ctx = With(ctx, prefix)

	ret, err := handler(ctx, req)
	duration := td()
	if err != nil {
		Infoc(ctx, "Failed to handle method %s handle_time=%s handle_err=%s", info.FullMethod, duration, err)
	} else {
		Infoc(ctx, "Succeed to handle method %s handle_time=%s", info.FullMethod, duration)
	}
	return ret, err
}

func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ss.Context()

	prefix := fmt.Sprintf("[%s]", strings.TrimPrefix(info.FullMethod, "/"))
	ctx = With(ctx, prefix)
	td := TimeFuncDuration()
	// Wrap a server stream implementation to modify the context to include the data.
	err := handler(srv, ss)
	duration := td()
	if err != nil {
		Infoc(ctx, "Failed to handle stream method %s handle_time=%s handle_err=%s", info.FullMethod, duration, err)
	} else {
		Infoc(ctx, "Succeed to handle stream method %s handle_time=%s", info.FullMethod, duration)
	}

	return err
}
