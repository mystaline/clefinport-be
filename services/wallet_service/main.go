package main

import (
	"log"
	"os"

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

	app := app.MakeApp()
	app.Run(&serviceProvider)
}
