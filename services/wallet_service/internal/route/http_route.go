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
	wallet := app.Group("/v1/wallet")

	// // Get wallet member list
	// wallet.Get("/:id/members", walletController.GetWalletMemberList)
	// // Get wallet latest 5 transaction list
	// wallet.Get("/:id/latest-transactions", walletController.GetWalletLatestTransactionList)
	// // Get all wallet transactions
	// wallet.Get("/:id/detail-transactions", walletController.GetWalletTransactions)
	// Get wallet detail
	wallet.Get("/:id", walletController.GetWalletInfo)
	// // Create new wallet
	// wallet.Post("", walletController.CreateWallet)
	// // Transfer between wallet
	// wallet.Post("/:id/transfer", walletController.TransferBalance)
	// // Invite member to shared wallet
	// wallet.Post("/:id/invite-member", walletController.InviteCollabMember)
	// // Accept invitation to shared wallet
	// wallet.Post("/:id/accept-invitation", walletController.AcceptCollabInvitation)
	// // Delete member from shared wallet
	// wallet.Delete("/:id/delete-member", walletController.DeleteMember)
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
