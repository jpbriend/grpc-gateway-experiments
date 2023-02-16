package internal

import (
	"context"
	"fmt"
	pb "github.com/jpbriend/grpc-gateway-experiments/generated/potato"
	"github.com/rs/zerolog/log"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

func newPotatoes() []*pb.Potato {
	return []*pb.Potato{
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
		{
			Id:   "3",
			Name: "Big Potato",
			Size: 666,
		},
		{
			Id:   "4",
			Name: "Summer Potato",
			Size: 10,
		},
	}
}

// GetPotatoes returns a list of potatoes
// Sorting is handle via the *order_by* field
// Pagination is handled via the *page_size* and *pagen* fields
func (gs *GRPCServer) GetPotatoes(ctx context.Context, req *pb.GetPotatoesRequest) (*pb.GetPotatoesResponse, error) {
	res := &pb.GetPotatoesResponse{
		Potatoes: newPotatoes(),
	}
	res, err := withOrder(res, req.OrderBy)
	if err != nil {
		log.Error().Err(err).Msg("Failed to sort")
		return nil, err
	}

	res, err = withPagination(res, int(req.PageSize), int(req.Page))
	if err != nil {
		return nil, err
	}
	return res, nil
}

// The following pagination is only for demo purpose. Real pagination should be done in the database.
func withPagination(p *pb.GetPotatoesResponse, pageSize int, page int) (*pb.GetPotatoesResponse, error) {
	log.Debug().Int("pageSize", pageSize).Int("page", page).Msg("with Pagination")

	// Check boundaries
	if pageSize < 0 {
		// TODO refactor to a dedicated function
		st := status.New(codes.InvalidArgument, "pageSize must be greater than 0")
		desc := "The page size must be positive"
		v := &errdetails.BadRequest_FieldViolation{
			Field:       "pageSize",
			Description: desc,
		}
		br := &errdetails.BadRequest{}
		br.FieldViolations = append(br.FieldViolations, v)
		st, err := st.WithDetails(br)
		if err != nil {
			// If this errored, it will always error here, so better panic so we can figure
			// out why than have this silently passing.
			panic(fmt.Sprintf("Unexpected error attaching metadata: %v", err))
		}
		return nil, st.Err()
	}

	if pageSize == 0 || pageSize >= len(p.Potatoes) {
		return p, nil
	}

	maxPage := len(p.Potatoes)/pageSize - 1
	var pt = 0

	if page != 0 {
		pt = page
	}

	if pt > maxPage {
		// Improve the error handling, cf a few lines above
		return nil, status.Errorf(codes.InvalidArgument, "page is out of bound")
	}

	// Compute paginated response
	if pageSize < len(p.Potatoes) {
		start := pageSize * pt
		end := pageSize * (pt + 1)
		p.Potatoes = p.Potatoes[start:end]
		if pt < maxPage {
			p.NextPage = int32(pt + 1)
		}
	}
	return p, nil
}

// The following sort is only for demo purpose. Real sorting should be done in the database.
func withOrder(p *pb.GetPotatoesResponse, orderBy string) (*pb.GetPotatoesResponse, error) {
	log.Debug().Str("orderBy", orderBy).Msg("with Sorting")
	if orderBy == "" {
		return p, nil
	}

	sorted, err := NewSorter[pb.Potato](orderBy).Sort(p.GetPotatoes())
	if err != nil {
		return nil, err
	} else {
		p.Potatoes = sorted
	}
	return p, nil
}
