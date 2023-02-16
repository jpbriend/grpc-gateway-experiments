package main

import (
	"flag"
	"github.com/jpbriend/grpc-gateway-experiments/internal"
)

func main() {
	config := internal.Configuration{}

	// Server config
	flag.IntVar(&config.ListenPort, "listenPort", 8081, "port the server is listening on")
	flag.BoolVar(&config.Debug, "debug", true, "Are DEBUG logs activated?")
	flag.BoolVar(&config.DevMode, "devMode", true, "Is devMode active? If true logs are pretty printed")

	// gRPC downstream services config
	flag.StringVar(&config.PotatoServiceEndpoint, "potato-endpoint", "localhost:8080", "URL of the potato-service")

	flag.Parse()

	internal.Start(&config)

}
