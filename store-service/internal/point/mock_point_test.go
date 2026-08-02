package point_test

import (
	"context"
	"store-service/internal/point"

	"github.com/stretchr/testify/mock"
)

type mockPointGateway struct {
	mock.Mock
}

func (gateway *mockPointGateway) GetPoints(ctx context.Context, userID int) ([]point.Point, error) {
	argument := gateway.Called(ctx, userID)
	return argument.Get(0).([]point.Point), argument.Error(1)
}

func (gateway *mockPointGateway) CreatePoint(ctx context.Context, userID int, pointItem point.Point) (point.Point, error) {
	argument := gateway.Called(ctx, userID, pointItem)
	return argument.Get(0).(point.Point), argument.Error(1)
}

func (gateway *mockPointGateway) CalculatePoint(ctx context.Context, priceThb float64) (int, error) {
	argument := gateway.Called(ctx, priceThb)
	return argument.Int(0), argument.Error(1)
}

func (gateway *mockPointGateway) ApproveEarnPoint(ctx context.Context, body point.ApproveEarnPointRequest) error {
	argument := gateway.Called(ctx, body)
	return argument.Error(0)
}

func (gateway *mockPointGateway) GetBalance(ctx context.Context, orgID int, uid int) ([]point.BalanceItem, error) {
	argument := gateway.Called(ctx, orgID, uid)
	return argument.Get(0).([]point.BalanceItem), argument.Error(1)
}
