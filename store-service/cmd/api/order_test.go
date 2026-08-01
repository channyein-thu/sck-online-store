package api_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"store-service/cmd/api"
	"store-service/internal/order"
	"store-service/internal/point"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockOrderService struct {
	mock.Mock
}

func (m *mockOrderService) CreateOrder(ctx context.Context, uid int, submitedOrder order.SubmitedOrder) (order.Order, error) {
	args := m.Called(ctx, uid, submitedOrder)
	return args.Get(0).(order.Order), args.Error(1)
}

func (m *mockOrderService) OrderBurnPoint(ctx context.Context, uid int, orderID int, burn int) error {
	args := m.Called(ctx, uid, orderID, burn)
	return args.Error(0)
}

func (m *mockOrderService) GetOrderSummary(ctx context.Context, orderNumber int64) (order.OrderSummary, error) {
	args := m.Called(ctx, orderNumber)
	return args.Get(0).(order.OrderSummary), args.Error(1)
}

func (m *mockOrderService) GeneratePDFFromData(orderDetail order.OrderSummary) ([]byte, error) {
	args := m.Called(orderDetail)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *mockOrderService) ListOrders(ctx context.Context, uid int) ([]order.OrderHistoryItem, error) {
	args := m.Called(ctx, uid)
	return args.Get(0).([]order.OrderHistoryItem), args.Error(1)
}

func (m *mockOrderService) ConfirmReceipt(ctx context.Context, uid int, orderNumber int64) error {
	args := m.Called(ctx, uid, orderNumber)
	return args.Error(0)
}

func newConfirmReceiptContext(uid int, orderNumberParam string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/api/v1/order/"+orderNumberParam+"/confirmReceipt", nil)
	ctx.Params = gin.Params{{Key: "id", Value: orderNumberParam}}
	ctx.Set("userID", uid)
	return ctx, recorder
}

func Test_ConfirmReceiptHandler_Input_Valid_OrderNumber_Should_Be_200(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001001

	mockService := new(mockOrderService)
	mockService.On("ConfirmReceipt", mock.Anything, uid, orderNumber).Return(nil)

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "2601069522001001")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusOK, recorder.Code)
}

func Test_ConfirmReceiptHandler_Input_Invalid_OrderNumber_Should_Be_400(t *testing.T) {
	uid := 1
	mockService := new(mockOrderService)

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "not-a-number")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusBadRequest, recorder.Code)
	mockService.AssertNotCalled(t, "ConfirmReceipt", mock.Anything, mock.Anything, mock.Anything)
}

func Test_ConfirmReceiptHandler_Input_ErrOrderNotFound_Should_Be_404(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001002

	mockService := new(mockOrderService)
	mockService.On("ConfirmReceipt", mock.Anything, uid, orderNumber).Return(order.ErrOrderNotFound)

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "2601069522001002")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func Test_ConfirmReceiptHandler_Input_ErrOrderNotOwned_Should_Be_403(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001003

	mockService := new(mockOrderService)
	mockService.On("ConfirmReceipt", mock.Anything, uid, orderNumber).Return(order.ErrOrderNotOwned)

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "2601069522001003")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
}

func Test_ConfirmReceiptHandler_Input_ErrNoPendingPoints_Should_Be_404(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001004

	mockService := new(mockOrderService)
	mockService.On("ConfirmReceipt", mock.Anything, uid, orderNumber).Return(point.ErrNoPendingPoints)

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "2601069522001004")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func Test_ConfirmReceiptHandler_Input_Unexpected_Error_Should_Be_500(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001005

	mockService := new(mockOrderService)
	mockService.On("ConfirmReceipt", mock.Anything, uid, orderNumber).Return(errors.New("db exploded"))

	orderAPI := api.OrderAPI{OrderService: mockService}
	ctx, recorder := newConfirmReceiptContext(uid, "2601069522001005")

	orderAPI.ConfirmReceiptHandler(ctx)

	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}
