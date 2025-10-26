package app

import (
	"os"

	"github.com/mystaline/clefinport-be/pkg/provider"

	wallet_route "github.com/mystaline/clefinport-be/services/wallet_service/internal/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"
)

type App struct {
	app *fiber.App
}

func MakeApp() *App {
	return &App{
		app: fiber.New(),
	}
}

func (a *App) Run(
	serviceProvider provider.IServiceProvider,
) {
	a.app.Use(cors.New())

	swaggerURL := "doc.json"
	env := os.Getenv("ENV")
	if env != "" {
		swaggerURL = "/TEMPLATE/docs/doc.json"
	}
	a.app.Get("/docs/*", swagger.New(swagger.Config{URL: swaggerURL}))

	setupRoute(a.app, serviceProvider)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	a.app.Listen(":" + port)
}

func setupRoute(
	app *fiber.App,
	serviceProvider provider.IServiceProvider,
) {
	// app.Use(util_middleware.ValidateJWTSQL())
	app.Use(logger.New())

	wallet_route.SetupWalletController(app, serviceProvider)
}
