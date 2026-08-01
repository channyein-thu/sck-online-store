package api

import (
	"log/slog"
	"net/http"

	"store-service/internal/point"

	"github.com/gin-gonic/gin"
)

type PointAPI struct {
	PointService point.PointInterface
}

// @Summary Get total points
// @Description Get user's total point balance
// @Tags point
// @Accept json
// @Produce json
// @Success 200 {object} point.Point
// @Failure 500
// @Router /api/v1/point [get]
func (api PointAPI) TotalPointHandler(context *gin.Context) {
	uid := context.GetInt("userID")

	ctx := context.Request.Context()
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
