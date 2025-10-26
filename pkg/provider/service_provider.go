package provider

import (
	"github.com/mystaline/clefinport-be/pkg/db"
	"github.com/mystaline/clefinport-be/pkg/service"

	"github.com/stretchr/testify/mock"
)

type IServiceProvider interface {
	MakeService(dbName db.DBName) service.PostgreSqlService
}

type ServiceProvider struct{}

func (m *ServiceProvider) MakeService(dbName db.DBName) service.PostgreSqlService {
	return service.MakeService(dbName)
}

type MockServiceProvider struct {
	mock.Mock
}

func (m *MockServiceProvider) MakeService(dbName db.DBName) service.PostgreSqlService {
	args := m.Called(dbName)
	return args.Error(0).(service.PostgreSqlService)
}
