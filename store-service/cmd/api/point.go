package api

import (
	"log/slog"
	"net/http"
	"strconv"

	"store-service/internal/point"

	"github.com/gin-gonic/gin"
)

type PointAPI struct {
	PointService point.PointInterface
}

// @Summary Deduct points from user
// @Description Deduct points from user's point balance
// @Tags point
// @Accept json
// @Produce json
// @Param request body point.SubmitedPoint true "Point deduction request"
// @Success 200 {object} point.Point
// @Failure 400 {string} string "Bad request error"
// @Failure 500
// @Router /api/v1/point [post]
func (api PointAPI) DeductPointHandler(context *gin.Context) {
	ctx := context.Request.Context()

	var request point.SubmitedPoint
	if err := context.BindJSON(&request); err != nil {
		slog.ErrorContext(ctx, "Point deduct bad request",
			"log_type", "error",
			"error_code", "INVALID_REQUEST",
			"error_message", err.Error(),
			"user_id", 0,
		)
		context.String(http.StatusBadRequest, err.Error())
		return
	}

	uid := context.GetInt("userID")

	res, err := api.PointService.DeductPoint(ctx, uid, request)
	if err != nil {
		slog.ErrorContext(ctx, "PointService.DeductPoint failed",
			"log_type", "error",
			"error_code", "POINT_DEDUCTION_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
			slog.Any("request", map[string]any{"amount": request.Amount}),
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	slog.InfoContext(ctx, "Points deducted",
		"log_type", "business",
		"event", "points_deducted",
		"entity_type", "point",
		"entity_id", uid,
		"actor_id", uid,
		slog.Any("metadata", map[string]any{
			"amount":          request.Amount,
			"remaining_point": res.Point,
		}),
	)

	context.JSON(http.StatusOK, res)
}

// @Summary Get point balance by status
// @Description Get user's point balance broken down by status (PENDING_APPROVAL, APPROVED, REDEEMED, EXPIRED)
// @Tags point
// @Accept json
// @Produce json
// @Success 200 {array} point.BalanceItem
// @Failure 500
// @Router /api/v1/point [get]
func (api PointAPI) TotalPointHandler(context *gin.Context) {
	uid := context.GetInt("userID")

	ctx := context.Request.Context()
	res, err := api.PointService.GetBalance(ctx, uid)

	if err != nil {
		slog.ErrorContext(ctx, "PointService.GetBalance failed",
			"log_type", "error",
			"error_code", "POINT_QUERY_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, res)
}

// @Summary Get total points for a specific user
// @Description Get a user's total point balance by user ID
// @Tags point
// @Produce json
// @Param userid path int true "User ID"
// @Success 200 {object} point.TotalPoint
// @Failure 400 {string} string "Bad request error"
// @Failure 500
// @Router /api/v1/point/{userid} [get]
func (api PointAPI) GetPointByUserIDHandler(context *gin.Context) {
	ctx := context.Request.Context()

	uid, err := strconv.Atoi(context.Param("userid"))
	if err != nil {
		slog.ErrorContext(ctx, "Get point by user ID bad request",
			"log_type", "error",
			"error_code", "INVALID_REQUEST",
			"error_message", err.Error(),
		)
		context.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := api.PointService.TotalPoint(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "PointService.TotalPoint failed",
			"log_type", "error",
			"error_code", "POINT_QUERY_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, res)
}

// @Summary Calculate points earned for a THB price
// @Description Calculate how many points would be earned for a given THB amount
// @Tags point
// @Produce json
// @Param priceThb query number true "Price in THB"
// @Success 200 {object} point.TotalPoint
// @Failure 400 {string} string "Bad request error"
// @Failure 500
// @Router /api/v1/point/calculate [get]
func (api PointAPI) CalculatePointHandler(context *gin.Context) {
	ctx := context.Request.Context()

	priceThb, err := strconv.ParseFloat(context.Query("priceThb"), 64)
	if err != nil {
		slog.ErrorContext(ctx, "Point calculate bad request",
			"log_type", "error",
			"error_code", "INVALID_REQUEST",
			"error_message", err.Error(),
		)
		context.String(http.StatusBadRequest, err.Error())
		return
	}

	res, err := api.PointService.CalculatePoint(ctx, priceThb)
	if err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, res)
}
