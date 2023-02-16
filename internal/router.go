package internal

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/jpbriend/grpc-gateway-experiments/generated/potato"
	"google.golang.org/grpc"
)

type Router struct{}

func (r *Router) RegisterPotatoService(ctx context.Context, m *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error {
	return pb.RegisterPotatoServiceHandlerFromEndpoint(ctx, m, endpoint, opts)
}
