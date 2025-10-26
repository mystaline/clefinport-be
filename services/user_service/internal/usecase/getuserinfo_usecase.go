package usecase

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/mystaline/clefinport-be/services/user_service/internal/dto"

	db "github.com/mystaline/clefinport-be/pkg/db"
	"github.com/mystaline/clefinport-be/pkg/entity"
	provider "github.com/mystaline/clefinport-be/pkg/provider"
	service "github.com/mystaline/clefinport-be/pkg/service"
	"github.com/mystaline/clefinport-be/pkg/sql_query"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
)

type GetUserInfoParam struct {
	Ctx    context.Context
	UserID string
}

type GetUserInfoUseCase struct {
	UserService service.PostgreSqlService

	ServiceProvider provider.IServiceProvider
	WalletClient    pb_wallet.WalletServiceClient
}

func MakeGetUserInfoUseCase(
	serviceProvider provider.IServiceProvider,
	walletClient pb_wallet.WalletServiceClient,
) *GetUserInfoUseCase {
	return &GetUserInfoUseCase{
		ServiceProvider: serviceProvider,
		WalletClient:    walletClient,
	}
}

func (u *GetUserInfoUseCase) InitService() {
	dbName := db.UserServiceDBName

	u.UserService = u.ServiceProvider.MakeService(dbName)
	u.UserService.Debug(2)
}

func (u *GetUserInfoUseCase) Invoke(
	param GetUserInfoParam,
) (*dto.GetUserInfoResult, error) {
	res, err := u.WalletClient.GetTotalBalanceByUserId(param.Ctx, &pb_wallet.GetTotalBalanceByUserIdRequest{
		UserId: param.UserID,
	})
	if err != nil {
		return nil, err
	}

	if res.UserId != param.UserID {
		return nil, &entity.HttpError{
			Code:    fiber.StatusInternalServerError,
			Message: "mismatch user id when receiving response from grpc server",
			Err:     err,
		}
	}

	query, args, _ := sql_query.
		NewSQLSelectBuilder[dto.GetUserInfoData](db.UserTableName).
		Where(map[string]sql_query.SQLCondition{
			"id": {Operator: sql_query.SQLOperatorEqual, Value: param.UserID},
		}).
		SetLimit(1).
		Build()

	var user dto.GetUserInfoResult
	if err := u.UserService.SelectOne(&user, param.Ctx, query, args...); err != nil {
		return nil, err
	}
	user.TotalBalance = int(res.TotalBalance)

	return &user, nil
}
