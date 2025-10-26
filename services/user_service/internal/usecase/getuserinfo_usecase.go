package usecase

import (
	"context"

	"github.com/mystaline/clefinport-be/services/user_service/internal/dto"

	db "github.com/mystaline/clefinport-be/pkg/db"
	provider "github.com/mystaline/clefinport-be/pkg/provider"
	service "github.com/mystaline/clefinport-be/pkg/service"
	"github.com/mystaline/clefinport-be/pkg/sql_query"
)

type GetUserInfoParam struct {
	Ctx    context.Context
	UserID string
}

type GetUserInfoUseCase struct {
	Service service.PostgreSqlService

	ServiceProvider provider.IServiceProvider
}

func MakeGetUserInfoUseCase(
	serviceProvider provider.IServiceProvider,
) *GetUserInfoUseCase {
	return &GetUserInfoUseCase{
		ServiceProvider: serviceProvider,
	}
}

func (u *GetUserInfoUseCase) InitService() {
	dbName := db.UserServiceDBName

	u.Service = u.ServiceProvider.MakeService(dbName)
	u.Service.Debug(2)
}

func (u *GetUserInfoUseCase) Invoke(
	param GetUserInfoParam,
) (*dto.GetUserInfoResult, error) {
	query, args, _ := sql_query.
		NewSQLSelectBuilder[dto.GetUserInfoData](db.UserTableName).
		Where(map[string]sql_query.SQLCondition{
			"id": {Operator: sql_query.SQLOperatorEqual, Value: param.UserID},
		}).
		SetLimit(1).
		Build()

	var user dto.GetUserInfoResult
	if err := u.Service.SelectOne(&user, param.Ctx, query, args...); err != nil {
		return nil, err
	}

	return &user, nil
}
