package route

import (
	"time"

	"github.com/mystaline/clefinport-be/services/wallet_service/internal/controller"
	"github.com/mystaline/clefinport-be/services/wallet_service/internal/usecase"

	"github.com/gofiber/fiber/v2"

	"github.com/mystaline/clefinport-be/pkg/provider"
)

func SetupWalletRoute(
	app *fiber.App,
	walletController controller.WalletController,
) {
	wallet := app.Group("/v1")

	wallet.Get("/wallet/:id", walletController.GetWalletInfo)
}

func SetupWalletController(
	app *fiber.App,
	serviceProvider provider.IServiceProvider,
) {
	getWalletInfoUsecase := usecase.MakeGetWalletInfoUseCase(serviceProvider)

	walletController := controller.MakeWalletController(
		60*time.Second,

		getWalletInfoUsecase,
	)

	SetupWalletRoute(app, *walletController)
}
