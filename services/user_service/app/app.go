package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/mystaline/clefinport-be/pkg/provider"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user_route "github.com/mystaline/clefinport-be/services/user_service/internal/route"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/swagger"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
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

	grpcHost := os.Getenv("WALLET_GRPC_HOST")
	grpcAddr := os.Getenv("WALLET_GRPC_ADDRESS")
	target := fmt.Sprintf("%s:%s", grpcHost, grpcAddr)
	conn := mustConnectGRPC(target, 10)

	startDial := time.Now()
	walletClient := pb_wallet.NewWalletServiceClient(conn)
	log.Println("Dial done in", time.Since(startDial))

	setupRoute(a.app, serviceProvider, walletClient)

	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		port = "8080"
	}

	a.app.Listen(":" + port)
}

func mustConnectGRPC(target string, retries int) *grpc.ClientConn {
	var conn *grpc.ClientConn
	var err error

	for i := 1; i <= retries; i++ {
		conn, err = grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			fmt.Println("✅ Connected to", target)
			return conn
		}
		fmt.Printf("⏳ Retry %d/%d connecting to %s: %v\n", i, retries, target, err)
		time.Sleep(2 * time.Second)
	}

	panic("❌ Failed to connect to gRPC service after retries: " + err.Error())
}

func setupRoute(
	app *fiber.App,
	serviceProvider provider.IServiceProvider,
	walletClient pb_wallet.WalletServiceClient,
) {
	// app.Use(util_middleware.ValidateJWTSQL())
	app.Use(logger.New())

	user_route.SetupUserController(app, serviceProvider, walletClient)
}
