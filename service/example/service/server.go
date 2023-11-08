package service

import (
	"context"

	"github.com/superwhys/goutils/service/example/examplepb"
)

type ExampleService struct {
	examplepb.UnimplementedExampleHelloServiceServer
}

func NewExampleService() *ExampleService {
	return &ExampleService{}
}

func (es *ExampleService) SayHello(ctx context.Context, in *examplepb.HelloRequest) (*examplepb.HelloResponse, error) {
	return &examplepb.HelloResponse{
		Message: "Hello " + in.Name,
	}, nil
}
