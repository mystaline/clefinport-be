package route

import (
	"time"

	"github.com/mystaline/clefinport-be/services/user_service/internal/controller"
	"github.com/mystaline/clefinport-be/services/user_service/internal/usecase"

	"github.com/gofiber/fiber/v2"

	"github.com/mystaline/clefinport-be/pkg/provider"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
)

func SetupUserRoute(
	app *fiber.App,
	userController controller.UserController,
) {
	dashboard := app.Group("/v1")

	dashboard.Get("/user/:id", userController.GetUserInfo)
}

func SetupUserController(
	app *fiber.App,
	serviceProvider provider.IServiceProvider,
	walletClient pb_wallet.WalletServiceClient,
) {
	getUserInfoUsecase := usecase.MakeGetUserInfoUseCase(serviceProvider, walletClient)

	userController := controller.MakeUserController(
		60*time.Second,

		getUserInfoUsecase,
	)

	SetupUserRoute(app, *userController)
}
