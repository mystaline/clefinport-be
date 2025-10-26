package controller

import (
	"context"

	"github.com/mystaline/clefinport-be/services/wallet_service/internal/dto"
	"github.com/mystaline/clefinport-be/services/wallet_service/internal/usecase"

	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/mystaline/clefinport-be/pkg/entity"
	"github.com/mystaline/clefinport-be/pkg/http"
)

type WalletController struct {
	Timeout time.Duration

	GetWalletInfoUsecase entity.UseCase[usecase.GetWalletInfoParam, *dto.GetWalletInfoResult]
}

func MakeWalletController(
	timeout time.Duration,

	getWalletInfoUseCase entity.UseCase[usecase.GetWalletInfoParam, *dto.GetWalletInfoResult],
) *WalletController {
	return &WalletController{
		Timeout:              timeout,
		GetWalletInfoUsecase: getWalletInfoUseCase,
	}
}

// @Summary      Get Wallet Info
// @Tags         Wallets
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} "Successfully get wallet info"
// @Router       /api/v1/wallet/:id [get]
func (c *WalletController) GetWalletInfo(ctx *fiber.Ctx) error {
	walletId := ctx.Params("id")

	return http.RunWithTimeout(
		ctx,
		c.Timeout,
		func(ctxWithTimeout context.Context) (*dto.GetWalletInfoResult, *entity.HttpError) {
			c.GetWalletInfoUsecase.InitService()

			param := usecase.GetWalletInfoParam{
				Ctx:      ctxWithTimeout,
				WalletID: walletId,
			}

			res, err := c.GetWalletInfoUsecase.Invoke(param)
			if err != nil {
				e := entity.ToHttpError(err)
				return nil, e
			}

			return res, nil
		}, "Successfully retrieve wallet info", fiber.StatusOK,
	)
}
