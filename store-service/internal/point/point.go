package point

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrNoPendingPoints = errors.New("no pending points to approve for this order")

type PointInterface interface {
	TotalPoint(ctx context.Context, uid int) (TotalPoint, error)
	DeductPoint(ctx context.Context, uid int, submitedPoint SubmitedPoint) (TotalPoint, error)
	CheckBurnPoint(ctx context.Context, uid int, amount int) (bool, error)
	CalculatePoint(ctx context.Context, priceThb float64) (TotalPoint, error)
	ApproveEarnPoint(ctx context.Context, uid int, orgID int, orderID int) error
	GetBalance(ctx context.Context, uid int) ([]BalanceItem, error)
}

type PointService struct {
	PointGateway PointGatewayInterface
}

type PointGatewayInterface interface {
	GetPoints(ctx context.Context, uid int) ([]Point, error)
	CreatePoint(ctx context.Context, uid int, body Point) (Point, error)
	CalculatePoint(ctx context.Context, priceThb float64) (int, error)
	ApproveEarnPoint(ctx context.Context, body ApproveEarnPointRequest) error
	GetBalance(ctx context.Context, orgID int, uid int) ([]BalanceItem, error)
}

func (pointService PointService) TotalPoint(ctx context.Context, uid int) (TotalPoint, error) {
	points, err := pointService.PointGateway.GetPoints(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.GetPoints failed",
			"log_type", "error", "error_code", "POINT_GATEWAY_FAILED", "error_message", err.Error(), "user_id", uid)
	}

	total := 0
	for _, point := range points {
		total += point.Amount
	}
	return TotalPoint{
		Point: total,
	}, err
}

func (pointService PointService) DeductPoint(ctx context.Context, uid int, submitedPoint SubmitedPoint) (TotalPoint, error) {
	_, err := pointService.CheckBurnPoint(ctx, uid, submitedPoint.Amount)
	if err != nil {
		return TotalPoint{}, err
	}

	point := Point{
		OrgID:   1,
		UserID:  uid,
		OrderID: submitedPoint.OrderID,
		Amount:  submitedPoint.Amount,
	}
	_, err_ := pointService.PointGateway.CreatePoint(ctx, uid, point)
	if err_ != nil {
		slog.ErrorContext(ctx, "PointGateway.CreatePoint failed",
			"log_type", "error", "error_code", "POINT_CREATE_FAILED", "error_message", err_.Error(),
			"user_id", uid, "order_id", submitedPoint.OrderID, "amount", submitedPoint.Amount)
		return TotalPoint{}, err_
	}
	return pointService.TotalPoint(ctx, uid)
}

func (pointService PointService) CalculatePoint(ctx context.Context, priceThb float64) (TotalPoint, error) {
	point, err := pointService.PointGateway.CalculatePoint(ctx, priceThb)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.CalculatePoint failed",
			"log_type", "error", "error_code", "POINT_CALCULATE_FAILED", "error_message", err.Error(), "price_thb", priceThb)
		return TotalPoint{}, err
	}
	return TotalPoint{Point: point}, nil
}

func (pointService PointService) ApproveEarnPoint(ctx context.Context, uid int, orgID int, orderID int) error {
	request := ApproveEarnPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
	}
	err := pointService.PointGateway.ApproveEarnPoint(ctx, request)
	if err != nil {
		if errors.Is(err, ErrNoPendingPoints) {
			slog.ErrorContext(ctx, "PointGateway.ApproveEarnPoint no pending points",
				"log_type", "error", "error_code", "POINT_APPROVE_NOT_FOUND", "error_message", err.Error(),
				"user_id", uid, "org_id", orgID, "order_id", orderID)
			return err
		}
		slog.ErrorContext(ctx, "PointGateway.ApproveEarnPoint failed",
			"log_type", "error", "error_code", "POINT_APPROVE_FAILED", "error_message", err.Error(),
			"user_id", uid, "org_id", orgID, "order_id", orderID)
		return err
	}
	return nil
}

func (pointService PointService) GetBalance(ctx context.Context, uid int) ([]BalanceItem, error) {
	balance, err := pointService.PointGateway.GetBalance(ctx, 1, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.GetBalance failed",
			"log_type", "error", "error_code", "POINT_BALANCE_FAILED", "error_message", err.Error(), "user_id", uid)
		return nil, err
	}
	return balance, nil
}

func (pointService PointService) CheckBurnPoint(ctx context.Context, uid int, amount int) (bool, error) {
	total, err := pointService.TotalPoint(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointService.TotalPoint failed",
			"log_type", "error", "error_code", "POINT_CHECK_FAILED", "error_message", err.Error(), "user_id", uid)
		return false, err
	}
	if amount+total.Point < 0 {
		return false, fmt.Errorf("points are not enough, please try again")
	}
	return true, nil
}
