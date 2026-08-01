package point_test

import (
	"context"
	"fmt"
	"store-service/internal/point"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_RedeemPoint_Input_Points_50_Should_Call_Gateway_No_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	points := 50

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("RedeemPoint", mock.Anything, point.RedeemPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
		Points:  points,
	}).Return(nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.RedeemPoint(context.Background(), uid, orgID, orderID, points)

	assert.Equal(t, nil, err)
	mockPointGateway.AssertExpectations(t)
}

func Test_RedeemPoint_Input_Gateway_ErrInsufficientPoints_Should_Return_ErrInsufficientPoints(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	points := 500

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("RedeemPoint", mock.Anything, point.RedeemPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
		Points:  points,
	}).Return(point.ErrInsufficientPoints)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.RedeemPoint(context.Background(), uid, orgID, orderID, points)

	assert.ErrorIs(t, err, point.ErrInsufficientPoints)
}

func Test_RedeemPoint_Input_Gateway_Error_Should_Return_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	points := 50
	expected := fmt.Errorf("response is not ok but it's 500")

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("RedeemPoint", mock.Anything, point.RedeemPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
		Points:  points,
	}).Return(expected)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.RedeemPoint(context.Background(), uid, orgID, orderID, points)

	assert.Equal(t, expected, err)
}

func Test_TotalPoint_Input_UserID_1_Should_Return_Balance_From_Gateway(t *testing.T) {
	expected := point.TotalPoint{
		Point: 150,
	}
	uid := 1
	orgID := 1

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("GetBalance", mock.Anything, orgID, uid).Return(150, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.TotalPoint(context.Background(), uid)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_TotalPoint_Input_Gateway_Error_Should_Return_Error(t *testing.T) {
	uid := 1
	orgID := 1
	expected := fmt.Errorf("response is not ok but it's 500")

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("GetBalance", mock.Anything, orgID, uid).Return(0, expected)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.TotalPoint(context.Background(), uid)

	assert.Equal(t, point.TotalPoint{}, actual)
	assert.Equal(t, expected, err)
}

func Test_EarnPoint_Input_AmountThb_10000_Should_Call_Gateway_With_Raw_Amount_No_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	amountThb := 10000.0

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CreateEarnPoint", mock.Anything, point.EarnPointRequest{
		OrgID:     orgID,
		UserID:    uid,
		OrderID:   orderID,
		AmountThb: amountThb,
	}).Return(nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.EarnPoint(context.Background(), uid, orgID, orderID, amountThb)

	assert.Equal(t, nil, err)
	mockPointGateway.AssertExpectations(t)
}

func Test_EarnPoint_Input_Gateway_Error_Should_Return_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	amountThb := 10000.0
	expected := fmt.Errorf("response is not ok but it's 500")

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CreateEarnPoint", mock.Anything, point.EarnPointRequest{
		OrgID:     orgID,
		UserID:    uid,
		OrderID:   orderID,
		AmountThb: amountThb,
	}).Return(expected)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.EarnPoint(context.Background(), uid, orgID, orderID, amountThb)

	assert.Equal(t, expected, err)
}

func Test_ApproveEarnPoint_Input_OrderID_10_Should_Call_Gateway_No_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("ApproveEarnPoint", mock.Anything, point.ApproveEarnPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
	}).Return(nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.ApproveEarnPoint(context.Background(), uid, orgID, orderID)

	assert.Equal(t, nil, err)
	mockPointGateway.AssertExpectations(t)
}

func Test_ApproveEarnPoint_Input_Gateway_ErrNoPendingPoints_Should_Return_ErrNoPendingPoints(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("ApproveEarnPoint", mock.Anything, point.ApproveEarnPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
	}).Return(point.ErrNoPendingPoints)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.ApproveEarnPoint(context.Background(), uid, orgID, orderID)

	assert.ErrorIs(t, err, point.ErrNoPendingPoints)
}

func Test_ApproveEarnPoint_Input_Gateway_Error_Should_Return_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 10
	expected := fmt.Errorf("response is not ok but it's 500")

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("ApproveEarnPoint", mock.Anything, point.ApproveEarnPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
	}).Return(expected)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	err := pointService.ApproveEarnPoint(context.Background(), uid, orgID, orderID)

	assert.Equal(t, expected, err)
}
