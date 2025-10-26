package controller

import (
	"context"

	"github.com/mystaline/clefinport-be/services/user_service/internal/dto"
	"github.com/mystaline/clefinport-be/services/user_service/internal/usecase"

	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/mystaline/clefinport-be/pkg/delivery"
	"github.com/mystaline/clefinport-be/pkg/entity"
)

type UserController struct {
	Timeout time.Duration

	GetUserInfoUsecase entity.UseCase[usecase.GetUserInfoParam, *dto.GetUserInfoResult]
}

func MakeUserController(
	timeout time.Duration,

	getUserInfoUseCase entity.UseCase[usecase.GetUserInfoParam, *dto.GetUserInfoResult],
) *UserController {
	return &UserController{
		Timeout:            timeout,
		GetUserInfoUsecase: getUserInfoUseCase,
	}
}

// @Summary      Get User Info
// @Tags         Users
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Success      200 {object} "Successfully get user info"
// @Router       /api/v1/user/:id [get]
func (c *UserController) GetUserInfo(ctx *fiber.Ctx) error {
	userId := ctx.Params("id")

	return delivery.RunHTTPWithTimeout(
		ctx,
		c.Timeout,
		func(ctxWithTimeout context.Context) (*dto.GetUserInfoResult, *entity.HttpError) {
			c.GetUserInfoUsecase.InitService()

			param := usecase.GetUserInfoParam{
				Ctx:    ctxWithTimeout,
				UserID: userId,
			}

			res, err := c.GetUserInfoUsecase.Invoke(param)
			if err != nil {
				e := entity.ToHttpError(err)
				return nil, e
			}

			return res, nil
		}, "Successfully retrieve user info", fiber.StatusOK,
	)
}
