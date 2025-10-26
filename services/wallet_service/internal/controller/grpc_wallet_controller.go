package controller

import (
	"context"
	"time"

	"github.com/mystaline/clefinport-be/pkg/delivery"
	"github.com/mystaline/clefinport-be/pkg/entity"
	"github.com/mystaline/clefinport-be/pkg/pb/wallet"
	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
	"github.com/mystaline/clefinport-be/services/wallet_service/internal/usecase"
)

type WalletServer struct {
	pb_wallet.UnimplementedWalletServiceServer

	Timeout time.Duration

	GetUserTotalBalanceUsecase entity.UseCase[usecase.GetUserTotalBalanceParam, *pb_wallet.GetTotalBalanceByUserIdResponse]
}

func NewWalletServer(
	timeout time.Duration,
	getUserTotalBalanceUseCase entity.UseCase[usecase.GetUserTotalBalanceParam, *pb_wallet.GetTotalBalanceByUserIdResponse],
) *WalletServer {
	return &WalletServer{
		Timeout:                    timeout,
		GetUserTotalBalanceUsecase: getUserTotalBalanceUseCase,
	}
}

// Example RPC implementation
func (s *WalletServer) GetTotalBalanceByUserId(
	ctx context.Context,
	req *pb_wallet.GetTotalBalanceByUserIdRequest,
) (*pb_wallet.GetTotalBalanceByUserIdResponse, error) {
	res, err := delivery.RunGRPCWithTimeout(
		ctx,
		s.Timeout,
		func(ctxWithTimeout context.Context) (*pb_wallet.GetTotalBalanceByUserIdResponse, *entity.HttpError) {
			s.GetUserTotalBalanceUsecase.InitService()

			param := usecase.GetUserTotalBalanceParam{
				Ctx:    ctxWithTimeout,
				UserID: req.UserId,
			}

			res, err := s.GetUserTotalBalanceUsecase.Invoke(param)
			if err != nil {
				e := entity.ToHttpError(err)
				return nil, e
			}

			return res, nil
		},
	)
	if err != nil {
		return nil, err
	}

	return res.(*wallet.GetTotalBalanceByUserIdResponse), nil
}
