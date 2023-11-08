package main

import (
	"context"
	"time"

	"github.com/superwhys/goutils/flags"
	"github.com/superwhys/goutils/lg"
	"github.com/superwhys/goutils/service"
	"github.com/superwhys/goutils/service/example/examplepb"
	exampleSrv "github.com/superwhys/goutils/service/example/service"
	"google.golang.org/grpc"
)

func main() {
	flags.Parse()

	grpcSrv := exampleSrv.NewExampleService()

	cs := service.NewYazlService(
		service.WithServiceName(flags.GetServiceName()),
		service.WithHTTPCORS(),
		service.WithPprof(),
		service.WithGRPC(func(srv *grpc.Server) {
			examplepb.RegisterExampleHelloServiceServer(srv, grpcSrv)
		}),
		service.WithNamedWorker("5sFunc", func(ctx context.Context) error {
			for range time.NewTicker(5 * time.Second).C {
				lg.Info("Name worker run")
			}
			return nil
		}),
		service.WithGRPCUI(),
	)

	if err := cs.ListenAndServer(0); err != nil {
		lg.PanicError(err)
	}
}
