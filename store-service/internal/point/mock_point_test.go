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

func (gateway *mockPointGateway) GetBalance(ctx context.Context, orgID int, uid int) (int, error) {
	argument := gateway.Called(ctx, orgID, uid)
	return argument.Int(0), argument.Error(1)
}

func (gateway *mockPointGateway) CreateEarnPoint(ctx context.Context, body point.EarnPointRequest) error {
	argument := gateway.Called(ctx, body)
	return argument.Error(0)
}

func (gateway *mockPointGateway) ApproveEarnPoint(ctx context.Context, body point.ApproveEarnPointRequest) error {
	argument := gateway.Called(ctx, body)
	return argument.Error(0)
}

func (gateway *mockPointGateway) RedeemPoint(ctx context.Context, body point.RedeemPointRequest) error {
	argument := gateway.Called(ctx, body)
	return argument.Error(0)
}
