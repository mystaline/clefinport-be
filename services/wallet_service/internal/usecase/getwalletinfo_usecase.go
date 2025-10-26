package usecase

import (
	"context"

	"github.com/mystaline/clefinport-be/services/wallet_service/internal/dto"

	db "github.com/mystaline/clefinport-be/pkg/db"
	provider "github.com/mystaline/clefinport-be/pkg/provider"
	service "github.com/mystaline/clefinport-be/pkg/service"
	"github.com/mystaline/clefinport-be/pkg/sql_query"
)

type GetWalletInfoParam struct {
	Ctx      context.Context
	WalletID string
}

type GetWalletInfoUseCase struct {
	Service service.PostgreSqlService

	ServiceProvider provider.IServiceProvider
}

func MakeGetWalletInfoUseCase(
	serviceProvider provider.IServiceProvider,
) *GetWalletInfoUseCase {
	return &GetWalletInfoUseCase{
		ServiceProvider: serviceProvider,
	}
}

func (u *GetWalletInfoUseCase) InitService() {
	dbName := db.WalletServiceDBName

	u.Service = u.ServiceProvider.MakeService(dbName)
	u.Service.Debug(2)
}

func (u *GetWalletInfoUseCase) Invoke(
	param GetWalletInfoParam,
) (*dto.GetWalletInfoResult, error) {
	query, args, _ := sql_query.
		NewSQLSelectBuilder[dto.GetWalletInfoData](db.WalletTableName).
		Where(map[string]sql_query.SQLCondition{
			"id": {Operator: sql_query.SQLOperatorEqual, Value: param.WalletID},
		}).
		SetLimit(1).
		Build()

	var wallet dto.GetWalletInfoResult
	if err := u.Service.SelectOne(&wallet, param.Ctx, query, args...); err != nil {
		return nil, err
	}

	return &wallet, nil
}
