package point

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var ErrNoPendingPoints = errors.New("no pending points to approve for this order")
var ErrInsufficientPoints = errors.New("insufficient points to redeem")

type PointInterface interface {
	TotalPoint(ctx context.Context, uid int) (TotalPoint, error)
	CheckBurnPoint(ctx context.Context, uid int, amount int) (bool, error)
	EarnPoint(ctx context.Context, uid int, orgID int, orderID int, amountThb float64) error
	ApproveEarnPoint(ctx context.Context, uid int, orgID int, orderID int) error
	RedeemPoint(ctx context.Context, uid int, orgID int, orderID int, points int) error
}

type PointService struct {
	PointGateway PointGatewayInterface
}

type PointGatewayInterface interface {
	GetPoints(ctx context.Context, uid int) ([]Point, error)
	GetBalance(ctx context.Context, orgID int, uid int) (int, error)
	CreateEarnPoint(ctx context.Context, body EarnPointRequest) error
	ApproveEarnPoint(ctx context.Context, body ApproveEarnPointRequest) error
	RedeemPoint(ctx context.Context, body RedeemPointRequest) error
}

func (pointService PointService) TotalPoint(ctx context.Context, uid int) (TotalPoint, error) {
	orgID := 1
	balance, err := pointService.PointGateway.GetBalance(ctx, orgID, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.GetBalance failed",
			"log_type", "error", "error_code", "POINT_GATEWAY_FAILED", "error_message", err.Error(), "user_id", uid)
		return TotalPoint{}, err
	}
	return TotalPoint{
		Point: balance,
	}, nil
}

func (pointService PointService) EarnPoint(ctx context.Context, uid int, orgID int, orderID int, amountThb float64) error {
	request := EarnPointRequest{
		OrgID:     orgID,
		UserID:    uid,
		OrderID:   orderID,
		AmountThb: amountThb,
	}
	err := pointService.PointGateway.CreateEarnPoint(ctx, request)
	if err != nil {
		slog.ErrorContext(ctx, "PointGateway.CreateEarnPoint failed",
			"log_type", "error", "error_code", "POINT_EARN_FAILED", "error_message", err.Error(),
			"user_id", uid, "order_id", orderID, "amount_thb", amountThb)
		return err
	}
	return nil
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

func (pointService PointService) RedeemPoint(ctx context.Context, uid int, orgID int, orderID int, points int) error {
	request := RedeemPointRequest{
		OrgID:   orgID,
		UserID:  uid,
		OrderID: orderID,
		Points:  points,
	}
	err := pointService.PointGateway.RedeemPoint(ctx, request)
	if err != nil {
		if errors.Is(err, ErrInsufficientPoints) {
			slog.ErrorContext(ctx, "PointGateway.RedeemPoint insufficient points",
				"log_type", "error", "error_code", "POINT_REDEEM_INSUFFICIENT", "error_message", err.Error(),
				"user_id", uid, "org_id", orgID, "order_id", orderID, "points", points)
			return err
		}
		slog.ErrorContext(ctx, "PointGateway.RedeemPoint failed",
			"log_type", "error", "error_code", "POINT_REDEEM_FAILED", "error_message", err.Error(),
			"user_id", uid, "org_id", orgID, "order_id", orderID, "points", points)
		return err
	}
	return nil
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
