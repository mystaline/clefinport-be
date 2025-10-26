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
	user := app.Group("/v1/user")

	// // Get user's wallet list
	// user.Get("/:id/wallets", userController.GetUserWalletList)
	// Get user info
	user.Get("/:id", userController.GetUserInfo)
	// // Change password
	// user.Put("/:id/password", userController.ChangePassword)
	// // Update profile
	// user.Put("/:id", userController.UpdateUserProfile)
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
