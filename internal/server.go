package internal

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jpbriend/grpc-gateway-experiments/internal/middleware"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"net/http"
	"os"
	"os/signal"
	"strconv"
)

type Configuration struct {
	ListenPort            int
	Debug                 bool
	DevMode               bool
	PotatoServiceEndpoint string
}

const (
	urlHealthcheck = "/z/health"
)

var (
	r *gin.Engine
)

func Start(c *Configuration) {
	middleware.ConfigureLogging(c.Debug, c.DevMode)
	log.Info().Msg(c.dumpGRPCConfiguration())

	r = gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.DefaultStructuredLogger())

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// Handle ctrl-C signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		os.Exit(2)
	}()

	// Start the downstream gRPC server
	NewGRPCServer().Start()

	registerGlobalEndpoints()
	registerGRPCHandlers(ctx, c)

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	log.Info().Msgf("Starting server on %v", c.ListenPort)
	if err := r.Run(":" + strconv.Itoa(c.ListenPort)); err != nil {
		log.Fatal().Err(err).Msg("could not run server")
	}
}

// Register gRPC server endpoint
// Note: Make sure the gRPC server is running properly and accessible
func registerGRPCHandlers(ctx context.Context, c *Configuration) {
	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	mux := runtime.NewServeMux(
		// Use in order to leverage the omitempty flag in the JSON definitions
		// to remove empty fields from the response
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}))
	rt := Router{}
	err := rt.RegisterPotatoService(ctx, mux, c.PotatoServiceEndpoint, opts)
	if err != nil {
		log.Error().Err(err).Msg("error when registering potato endpoint")
	}

	r.Group("/v1/*{grpc_gateway}").Any("", gin.WrapH(mux))
}

// healthzServer returns a simple health handler which returns ok.
func registerGlobalEndpoints() {
	// Register a Healthcheck endpoint
	r.GET(urlHealthcheck, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "OK",
		})
	})
}

// dumpGRPCConfiguration writes downstream gRPC endpoints to logs
func (c *Configuration) dumpGRPCConfiguration() string {
	res := "Endpoints:\n"
	res = res + fmt.Sprintf("  * potato-service: %s\n", c.PotatoServiceEndpoint)
	return res
}
