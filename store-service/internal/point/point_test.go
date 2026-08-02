package point_test

import (
	"context"
	"fmt"
	"store-service/internal/point"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_DeductPoint_Input_Amount_100_Should_be_Point_100(t *testing.T) {
	expected := point.TotalPoint{
		Point: 100,
	}
	uid := 1
	pointItem := point.Point{
		OrgID:  1,
		UserID: uid,
		Amount: 100,
	}
	pointList := []point.Point{
		pointItem,
	}

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CreatePoint", mock.Anything, uid, pointItem).Return(pointItem, nil)
	mockPointGateway.On("GetPoints", mock.Anything, uid).Return(pointList, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.DeductPoint(context.Background(), uid, point.SubmitedPoint{
		Amount: 100,
	})

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_DeductPoint_Input_Amount_Minus_100_Should_be_Error(t *testing.T) {
	expected := fmt.Errorf("points are not enough, please try again")
	uid := 1
	pointItem := point.Point{
		OrgID:  1,
		UserID: uid,
		Amount: -100,
	}
	pointList := []point.Point{
		pointItem,
	}

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CreatePoint", mock.Anything, uid, pointItem).Return(pointItem, nil)
	mockPointGateway.On("GetPoints", mock.Anything, uid).Return(pointList, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	_, err := pointService.DeductPoint(context.Background(), uid, point.SubmitedPoint{
		Amount: -100,
	})

	assert.Equal(t, expected, err)
}

func Test_TotalPoint_Point_100_and_50_Should_be_Point_150(t *testing.T) {
	expected := point.TotalPoint{
		Point: 150,
	}
	uid := 1
	res := []point.Point{
		{
			OrgID:  1,
			UserID: 1,
			Amount: 100,
		},
		{
			OrgID:  1,
			UserID: 1,
			Amount: 50,
		},
	}

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("GetPoints", mock.Anything, uid).Return(res, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.TotalPoint(context.Background(), uid)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_CalculatePoint_Input_PriceThb_465_81_Should_be_Point_4(t *testing.T) {
	expected := point.TotalPoint{
		Point: 4,
	}
	priceThb := 465.81

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CalculatePoint", mock.Anything, priceThb).Return(4, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.CalculatePoint(context.Background(), priceThb)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_CalculatePoint_Should_be_Return_Error_When_PointGateway_Fails(t *testing.T) {
	expected := point.TotalPoint{}
	priceThb := 465.81

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("CalculatePoint", mock.Anything, priceThb).Return(0, fmt.Errorf("point-service unavailable"))

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.CalculatePoint(context.Background(), priceThb)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_TotalPoint_Point_100_and_Minus_50_Should_be_Point_50(t *testing.T) {
	expected := point.TotalPoint{
		Point: 50,
	}
	uid := 1
	res := []point.Point{
		{
			OrgID:  1,
			UserID: 1,
			Amount: 100,
		},
		{
			OrgID:  1,
			UserID: 1,
			Amount: -50,
		},
	}

	mockPointGateway := new(mockPointGateway)
	mockPointGateway.On("GetPoints", mock.Anything, uid).Return(res, nil)

	pointService := point.PointService{
		PointGateway: mockPointGateway,
	}
	actual, err := pointService.TotalPoint(context.Background(), uid)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}
