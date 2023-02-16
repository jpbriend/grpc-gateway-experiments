package internal

import (
	"context"
	pb "github.com/jpbriend/grpc-gateway-experiments/generated/potato"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"net"
)

type GRPCServer struct {
	pb.UnimplementedPotatoServiceServer
}

func NewGRPCServer() *GRPCServer {
	return &GRPCServer{}
}

func (gs *GRPCServer) Start() {
	lis, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen for gRPCServer")
	}

	s := grpc.NewServer()

	pb.RegisterPotatoServiceServer(s, gs)
	log.Info().Msg("Starting gRPC server on port 8080")

	go func() {
		log.Fatal().Err(s.Serve(lis)).Msg("Failed to serve gRPCServer")
	}()
}

func (gs *GRPCServer) GetPotato(ctx context.Context, req *pb.GetPotatoRequest) (*pb.GetPotatoResponse, error) {
	return &pb.GetPotatoResponse{
		Potato: &pb.Potato{
			Id:   req.PotatoId,
			Name: "Potato",
		},
	}, nil
}

var getPotatoesResponse = &pb.GetPotatoesResponse{
	Potatoes: []*pb.Potato{
		{
			Id:   "1",
			Name: "Potato 456",
			Size: 1,
		},
		{
			Id:   "2",
			Name: "Potato 123",
			Size: 42,
		},
	},
}

// GetPotatoes returns a list of potatoes
// Sorting is handle via the *order_by* field
func (gs *GRPCServer) GetPotatoes(ctx context.Context, req *pb.GetPotatoesRequest) (*pb.GetPotatoesResponse, error) {
	if req.OrderBy != "" {
		// The following sort is only for demo purpose. Real sorting should be done in the database.
		sorted, err := NewSorter[pb.Potato](req.OrderBy).Sort((*getPotatoesResponse).Potatoes)
		if err != nil {
			log.Error().Err(err).Msg("Failed to sort")
		}
		return &pb.GetPotatoesResponse{
			Potatoes: sorted}, nil
	}

	return getPotatoesResponse, nil
}
