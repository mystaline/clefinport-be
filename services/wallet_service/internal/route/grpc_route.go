package route

import (
	"time"

	"github.com/mystaline/clefinport-be/services/wallet_service/internal/controller"
	"github.com/mystaline/clefinport-be/services/wallet_service/internal/usecase"

	"github.com/mystaline/clefinport-be/pkg/provider"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
)

func SetupWalletGRPC(
	serviceProvider provider.IServiceProvider,
) pb_wallet.WalletServiceServer {
	grpcGetUserTotalBalanceUsecase := usecase.MakeGetUserTotalBalanceUseCase(serviceProvider)

	return controller.NewWalletServer(
		60*time.Second,

		grpcGetUserTotalBalanceUsecase,
	)
}
