package service

import (
	"context"
	"fmt"
	"time"

	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service/finder"
	"google.golang.org/grpc"
)

func DialGrpc(service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	return DialGrpcWithTimeOut(10*time.Second, service, opts...)
}

func DialGrpcWithTimeOut(timeout time.Duration, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return DialGrpcWithContext(ctx, service, opts...)

}

func DialGrpcWithContext(ctx context.Context, service string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts = append(opts, grpc.WithBlock(), grpc.WithInsecure())
	return dialGrpcWithTagContext(ctx, service, "", opts...)
}

func dialGrpcWithTagContext(ctx context.Context, service, tag string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := append(opts, grpc.WithBlock(), grpc.WithInsecure())
	options = append(options, opts...)

	address := finder.GetServiceFinder().GetAddressWithTag(service, tag)

	conn, err := grpc.DialContext(
		ctx,
		address,
		options...,
	)

	lg.Debug(fmt.Sprintf("dial grpc service %s with tag %s", service, tag))
	return conn, err
}
