package api

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"store-service/internal/order"
	"store-service/internal/point"

	"github.com/gin-gonic/gin"
)

// @title Store Service API
// @version 1.0
// @description Store service API documentation
// @host localhost:8000
// @BasePath /api/v1
type OrderAPI struct {
	OrderService order.OrderInterface
}

// OrderConfirmation represents the response after order submission
// @Description Order confirmation response containing order number
type OrderConfirmation struct {
	OrderNumber int64 `json:"order_number"`
}

// @Summary Submit new order
// @Description Creates a new order from the submitted order details
// @Tags order
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param order body order.SubmitedOrder true "Order details"
// @Success 200 {object} OrderConfirmation "Successfully created order"
// @Failure 400 {string} string "Bad Request - Invalid input"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500
// @Router /api/v1/order [post]
func (api OrderAPI) SubmitOrderHandler(context *gin.Context) {
	uid := context.GetInt("userID")
	ctx := context.Request.Context()

	var request order.SubmitedOrder
	if err := context.BindJSON(&request); err != nil {
		slog.ErrorContext(ctx, "Order submit bad request",
			"log_type", "error",
			"error_code", "INVALID_REQUEST",
			"error_message", err.Error(),
			"user_id", uid,
		)
		context.String(http.StatusBadRequest, err.Error())
		return
	}

	createdOrder, err := api.OrderService.CreateOrder(ctx, uid, request)
	if err != nil {
		slog.ErrorContext(ctx, "OrderService.CreateOrder failed",
			"log_type", "error",
			"error_code", "ORDER_CREATION_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
			slog.Any("request", map[string]any{
				"item_count":         len(request.Cart),
				"payment_method_id":  request.PaymentMethodID,
				"shipping_method_id": request.ShippingMethodID,
				"burn_point":         request.BurnPoint,
			}),
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	slog.InfoContext(ctx, "Order created",
		"log_type", "business",
		"event", "order_created",
		"entity_type", "order",
		"entity_id", createdOrder.OrderNumber,
		"actor_id", uid,
		slog.Any("metadata", map[string]any{
			"item_count":         len(request.Cart),
			"payment_method_id":  request.PaymentMethodID,
			"shipping_method_id": request.ShippingMethodID,
			"total_price":        request.TotalPrice,
			"burn_point":         request.BurnPoint,
		}),
	)

	context.JSON(http.StatusOK, OrderConfirmation{
		OrderNumber: createdOrder.OrderNumber,
	})
}

// @Summary Get order history
// @Description Returns the authenticated user's own orders, newest first
// @Tags order
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} order.OrderHistoryItem
// @Failure 401 {string} string "Unauthorized"
// @Failure 500
// @Router /api/v1/order/history [get]
func (api OrderAPI) GetOrderHistoryHandler(context *gin.Context) {
	uid := context.GetInt("userID")
	ctx := context.Request.Context()

	orders, err := api.OrderService.ListOrders(ctx, uid)
	if err != nil {
		slog.ErrorContext(ctx, "OrderService.ListOrders failed",
			"log_type", "error",
			"error_code", "ORDER_HISTORY_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	context.JSON(http.StatusOK, orders)
}

// @Summary Confirm receipt
// @Description Confirms the order was received: approves its earned points, starting the 180-day validity window
// @Tags order
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path int true "Order number"
// @Success 200
// @Failure 400 {string} string "Bad Request - Invalid order number"
// @Failure 401 {string} string "Unauthorized"
// @Failure 403 {object} map[string]string "Forbidden - order does not belong to this user"
// @Failure 404 {object} map[string]string "Not Found"
// @Failure 500
// @Router /api/v1/order/{id}/confirmReceipt [post]
func (api OrderAPI) ConfirmReceiptHandler(context *gin.Context) {
	uid := context.GetInt("userID")
	ctx := context.Request.Context()

	orderNumberStr := context.Param("id")
	orderNumber, err := strconv.ParseInt(orderNumberStr, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order number",
		})
		return
	}

	err = api.OrderService.ConfirmReceipt(ctx, uid, orderNumber)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			slog.ErrorContext(ctx, "OrderService.ConfirmReceipt not found",
				"log_type", "error",
				"error_code", "ORDER_NOT_FOUND",
				"error_message", err.Error(),
				"user_id", uid,
				slog.Any("request", map[string]any{"order_number": orderNumber}),
			)
			context.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, order.ErrOrderNotOwned) {
			slog.ErrorContext(ctx, "OrderService.ConfirmReceipt not owned",
				"log_type", "error",
				"error_code", "ORDER_NOT_OWNED",
				"error_message", err.Error(),
				"user_id", uid,
				slog.Any("request", map[string]any{"order_number": orderNumber}),
			)
			context.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
			return
		}
		if errors.Is(err, point.ErrNoPendingPoints) {
			slog.ErrorContext(ctx, "OrderService.ConfirmReceipt no pending points",
				"log_type", "error",
				"error_code", "POINT_APPROVE_NOT_FOUND",
				"error_message", err.Error(),
				"user_id", uid,
				slog.Any("request", map[string]any{"order_number": orderNumber}),
			)
			context.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		slog.ErrorContext(ctx, "OrderService.ConfirmReceipt internal error",
			"log_type", "error",
			"error_code", "ORDER_CONFIRM_RECEIPT_FAILED",
			"error_message", err.Error(),
			"user_id", uid,
			slog.Any("request", map[string]any{"order_number": orderNumber}),
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	slog.InfoContext(ctx, "Order receipt confirmed",
		"log_type", "business",
		"event", "order_receipt_confirmed",
		"entity_type", "order",
		"entity_id", orderNumber,
		"actor_id", uid,
	)

	context.JSON(http.StatusOK, gin.H{
		"order_number": orderNumber,
		"status":       "confirmed",
	})
}

func (api OrderAPI) GetOrderSummaryHandler(context *gin.Context) {
	acceptHeader := context.GetHeader("Accept")
	allowedHeaders := []string{"application/pdf", "", "*/*", "application/json"}
	isAllowed := false
	for _, header := range allowedHeaders {
		if header == acceptHeader {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		// 406 Not Acceptable
		context.AbortWithStatus(http.StatusNotAcceptable)
		return
	}

	ctx := context.Request.Context()
	orderNumberStr := context.Param("id")
	orderNumber, err := strconv.ParseInt(orderNumberStr, 10, 64)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid order number",
		})
		return
	}

	orderSummary, err := api.OrderService.GetOrderSummary(ctx, orderNumber)
	if err != nil {
		if errors.Is(err, order.ErrOrderNotFound) {
			slog.ErrorContext(ctx, "OrderService.GetOrderSummary not found",
				"log_type", "error",
				"error_code", "ORDER_NOT_FOUND",
				"error_message", err.Error(),
				"user_id", 0,
				slog.Any("request", map[string]any{"order_number": orderNumber}),
			)
			context.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		slog.ErrorContext(ctx, "OrderService.GetOrderSummary internal error",
			"log_type", "error",
			"error_code", "ORDER_SUMMARY_FAILED",
			"error_message", err.Error(),
			"user_id", 0,
			slog.Any("request", map[string]any{"order_number": orderNumber}),
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	if acceptHeader == "application/json" {
		context.JSON(http.StatusOK, orderSummary)
		return
	}

	pdfData, err := api.OrderService.GeneratePDFFromData(orderSummary)
	if err != nil {
		slog.ErrorContext(ctx, "OrderService.GeneratePDFFromData failed",
			"log_type", "error",
			"error_code", "PDF_GENERATION_FAILED",
			"error_message", err.Error(),
			"user_id", 0,
			slog.Any("request", map[string]any{"order_number": orderNumber}),
		)
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	filename := fmt.Sprintf("Order_Summary_%d.pdf", orderNumber)
	contentDisposition := fmt.Sprintf("attachment; filename=\"%s\"", filename)

	context.Header("Access-Control-Expose-Headers", "Content-Disposition")
	context.Header(
		"Content-Disposition",
		contentDisposition,
	)

	context.Data(http.StatusOK, "application/pdf", pdfData)
}
