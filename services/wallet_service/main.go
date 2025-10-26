package main

import (
	"log"
	"os"
	"sync"

	"github.com/mystaline/clefinport-be/pkg/provider"

	"github.com/joho/godotenv"

	"github.com/mystaline/clefinport-be/services/wallet_service/app"
)

func main() {
	if os.Getenv("DOCKER_ENV") == "" {
		if err := godotenv.Load(); err != nil {
			log.Println("No .env file found, using environment variables only")
		}
	}

	serviceProvider := provider.ServiceProvider{}

	var wg sync.WaitGroup
	wg.Add(2)

	// Start HTTP server
	go func() {
		defer wg.Done()

		app := app.MakeApp()
		app.Run(&serviceProvider)
	}()

	// Start gRPC server
	go func() {
		defer wg.Done()
		if err := app.RunGRPCServer(&serviceProvider); err != nil {
			log.Fatalf("failed to run grpc server: %v", err)
		}
	}()

	wg.Wait()
}
