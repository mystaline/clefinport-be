package usecase

import (
	"context"

	db "github.com/mystaline/clefinport-be/pkg/db"
	provider "github.com/mystaline/clefinport-be/pkg/provider"
	service "github.com/mystaline/clefinport-be/pkg/service"
	"github.com/mystaline/clefinport-be/pkg/sql_query"

	pb_wallet "github.com/mystaline/clefinport-be/pkg/pb/wallet"
)

type GetUserTotalBalanceParam struct {
	Ctx    context.Context
	UserID string
}

type GetUserTotalBalanceUseCase struct {
	Service service.PostgreSqlService

	ServiceProvider provider.IServiceProvider
}

func MakeGetUserTotalBalanceUseCase(
	serviceProvider provider.IServiceProvider,
) *GetUserTotalBalanceUseCase {
	return &GetUserTotalBalanceUseCase{
		ServiceProvider: serviceProvider,
	}
}

func (u *GetUserTotalBalanceUseCase) InitService() {
	dbName := db.WalletServiceDBName

	u.Service = u.ServiceProvider.MakeService(dbName)
	u.Service.Debug(2)
}

func (u *GetUserTotalBalanceUseCase) Invoke(
	param GetUserTotalBalanceParam,
) (*pb_wallet.GetTotalBalanceByUserIdResponse, error) {
	query, args, _ := sql_query.
		NewSQLSelectBuilder[any](db.UserWalletTableName).
		Select(`sum(balance) as balance`).
		Where(map[string]sql_query.SQLCondition{
			"user_id": {Operator: sql_query.SQLOperatorEqual, Value: param.UserID},
		}).
		Build()

	var wallet pb_wallet.GetTotalBalanceByUserIdResponse
	if err := u.Service.SelectOne(&wallet, param.Ctx, query, args...); err != nil {
		return nil, err
	}

	wallet.UserId = param.UserID

	return &wallet, nil
}
