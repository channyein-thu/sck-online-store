package order_test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"store-service/internal/auth"
	"store-service/internal/order"
	"store-service/internal/point"
	"store-service/internal/product"
	"store-service/internal/shipping"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test_CreateOrder_Input_Submitted_Order_Should_be_OrderNumber_2601069522001001(t *testing.T) {
	uid := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001001
	productPrice := 12.95
	nextSeq := 1
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	datePrefix := "260106"

	expected := order.Order{
		OrderNumber: orderNumber,
	}

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, submittedOrder.BurnPoint).Return(true, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, nil)

	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, submittedOrder.Cart[0].ProductID, submittedOrder.Cart[0].Quantity, productPrice).Return(nil)

	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("DeleteCart", mock.Anything, uid, submittedOrder.Cart[0].ProductID).Return(nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     mockCartRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
}

func Test_CreateOrder_Input_Submitted_Order_With_DiscountPrice_Should_Apply_Discount_Without_Double_Conversion(t *testing.T) {
	uid := 1
	oid := 8004359104
	var orderNumber int64 = 2601069522001002
	productPrice := 12.95
	nextSeq := 2
	fixedTime := time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC)
	datePrefix := "260106"
	burnPoint := 200

	expected := order.Order{
		OrderNumber: orderNumber,
	}

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            burnPoint,
		SubTotalPrice:        100.00,
		DiscountPrice:        100, // already in THB, from floor(burnPoint/2)
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -burnPoint).Return(true, nil)
	mockPointInterface.On("RedeemPoint", mock.Anything, uid, 1, oid, burnPoint).Return(nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	// subtotalPriceTHB (465.811034) minus the discount, which must be taken at face
	// value in THB and NOT run through ConvertToThb a second time
	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    100,
		TotalPrice:       415.811034,
		ShippingFee:      50,
		BurnPoint:        burnPoint,
		EarnPoint:        7,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, nil)

	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, submittedOrder.Cart[0].ProductID, submittedOrder.Cart[0].Quantity, productPrice).Return(nil)

	mockCartRepository := new(mockCartRepository)
	mockCartRepository.On("DeleteCart", mock.Anything, uid, submittedOrder.Cart[0].ProductID).Return(nil)

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		CartRepository:     mockCartRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, nil, err)
	assert.GreaterOrEqual(t, orderDetail.TotalPrice, 0.0)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Error_Points_not_Enough(t *testing.T) {
	expected := order.Order{}
	expectedErr := fmt.Errorf("points are not enough, please try again")

	uid := 1
	burnPoint := 100

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, -(burnPoint)).Return(false, fmt.Errorf("points are not enough, please try again"))

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            burnPoint,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, expectedErr, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_No_Product_in_Order_Error(t *testing.T) {
	expected := order.Order{}
	expectedErr := fmt.Errorf("There is no product in order, please try again")

	uid := 1

	submittedOrder := order.SubmitedOrder{
		Cart:                 []order.OrderProduct{},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.Equal(t, expectedErr, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Order_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "260112"
	nextSeq := 32
	fixedTime := time.Date(2026, 1, 12, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2601129522001032

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}
	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, errors.New("CreateOrder Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Shipping_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "261212"
	nextSeq := 80
	fixedTime := time.Date(2026, 12, 12, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2612129522001080

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, 0).Return(true, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, errors.New("CreateShipping Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_CreateOrder_Input_Submitted_Order_Should_be_Return_Create_Order_Product_Error(t *testing.T) {
	expected := order.Order{}

	uid := 1
	oid := 8004359103
	productPrice := 12.95
	datePrefix := "260515"
	nextSeq := 179
	fixedTime := time.Date(2026, 05, 15, 0, 0, 0, 0, time.UTC)
	var orderNumber int64 = 2605159522001179

	submittedOrder := order.SubmitedOrder{
		Cart: []order.OrderProduct{
			{
				ProductID: 2,
				Quantity:  1,
			},
		},
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
		PaymentMethodID:      1,
		BurnPoint:            0,
		SubTotalPrice:        100.00,
		DiscountPrice:        0,
		TotalPrice:           100.00,
	}

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("CheckBurnPoint", mock.Anything, uid, submittedOrder.BurnPoint).Return(true, nil)

	mockProductRepository := new(mockProductRepository)
	mockProductRepository.On("GetProductByID", mock.Anything, submittedOrder.Cart[0].ProductID).Return(product.ProductDetail{
		ID:           submittedOrder.Cart[0].ProductID,
		Name:         "43 Piece dinner Set",
		Price:        productPrice,
		PriceTHB:     0,
		PriceFullTHB: 0,
		Stock:        1,
		Brand:        "Coolkidz",
		Image:        "43_Piece_Dinner_Set.jpg",
	}, nil)

	mockShippingRepository := new(mockShippingRepository)
	mockShippingRepository.On("GetShippingMethodByID", mock.Anything, submittedOrder.ShippingMethodID).Return(shipping.ShippingMethodDetail{
		ID:          1,
		Name:        "Kerry",
		Description: "",
		Fee:         50,
	}, nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetNextSequence", mock.Anything, datePrefix, uid).Return(nextSeq, nil)

	mockOrderHelper := new(mockOrderHelper)
	mockOrderHelper.On("GenerateOrderNumber", submittedOrder.PaymentMethodID, submittedOrder.ShippingMethodID, uid, nextSeq, fixedTime).Return(orderNumber, nil)

	orderDetail := order.OrderDetail{
		OrderNumber:      orderNumber,
		ShippingMethodID: submittedOrder.ShippingMethodID,
		PaymentMethodID:  submittedOrder.PaymentMethodID,
		SubTotalPrice:    465.811034,
		DiscountPrice:    0,
		TotalPrice:       515.8110340000001,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        9,
	}

	mockOrderRepository.On("CreateOrder", mock.Anything, uid, orderDetail).Return(oid, nil)

	shippingInfo := order.ShippingInfo{
		ShippingMethodID:     1,
		ShippingAddress:      "405/37 ถ.มหิดล",
		ShippingSubDistrict:  "ท่าศาลา",
		ShippingDistrict:     "เมือง",
		ShippingProvince:     "เชียงใหม่",
		ShippingZipCode:      "50000",
		RecipientFirstName:   "ณัฐญา",
		RecipientLastName:    "ชุติบุตร",
		RecipientPhoneNumber: "0970809292",
	}
	mockOrderRepository.On("CreateShipping", mock.Anything, uid, oid, shippingInfo).Return(1, nil)

	mockOrderRepository.On("CreateOrderProduct", mock.Anything, oid, submittedOrder.Cart[0].ProductID, submittedOrder.Cart[0].Quantity, productPrice).Return(errors.New("CreateOrderProduct Error"))

	orderService := order.OrderService{
		ProductRepository:  mockProductRepository,
		OrderRepository:    mockOrderRepository,
		PointService:       mockPointInterface,
		ShippingRepository: mockShippingRepository,
		OrderHelper:        mockOrderHelper,
		Clock:              func() time.Time { return fixedTime },
	}

	actual, err := orderService.CreateOrder(context.Background(), uid, submittedOrder)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_OrderBurnPoint_Input_Burn_Points_100_Should_Call_RedeemPoint_No_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 8004359103
	burnPoint := 100

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("RedeemPoint", mock.Anything, uid, orgID, orderID, burnPoint).Return(nil)

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	err := orderService.OrderBurnPoint(context.Background(), uid, orderID, burnPoint)

	assert.Equal(t, nil, err)
}

func Test_OrderBurnPoint_Input_Burn_Points_100_Should_Return_RedeemPoint_Error(t *testing.T) {
	uid := 1
	orgID := 1
	orderID := 8004359103
	burnPoint := 100

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("RedeemPoint", mock.Anything, uid, orgID, orderID, burnPoint).Return(errors.New("RedeemPoint Error"))

	orderService := order.OrderService{
		PointService: mockPointInterface,
	}

	err := orderService.OrderBurnPoint(context.Background(), uid, orderID, burnPoint)

	assert.NotNil(t, err)
}

func Test_ListOrders_Input_UserID_4_Should_Return_Orders_Newest_First(t *testing.T) {
	userID := 4
	updatedTime := time.Date(2026, 2, 28, 18, 58, 44, 0, time.UTC)

	expected := []order.OrderHistoryItem{
		{
			OrderNumber:    2601069522002002,
			Status:         "paid",
			SubTotalPrice:  5246.22,
			TotalPrice:     5256.22,
			BurnPoint:      0,
			EarnPoint:      52,
			TrackingNumber: "KR-304590466",
			Updated:        updatedTime,
		},
		{
			OrderNumber:    2601069522001001,
			Status:         "created",
			SubTotalPrice:  4314.6,
			TotalPrice:     4364.6,
			BurnPoint:      0,
			EarnPoint:      43,
			TrackingNumber: "",
			Updated:        updatedTime,
		},
	}

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("ListOrdersByUserID", mock.Anything, userID).Return(expected, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
	}

	actual, err := orderService.ListOrders(context.Background(), userID)

	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}

func Test_ListOrders_Input_UserID_4_Should_Return_Repository_Error(t *testing.T) {
	userID := 4
	expected := []order.OrderHistoryItem(nil)

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("ListOrdersByUserID", mock.Anything, userID).Return(expected, errors.New("ListOrdersByUserID Error"))

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
	}

	actual, err := orderService.ListOrders(context.Background(), userID)

	assert.Equal(t, expected, actual)
	assert.NotNil(t, err)
}

func Test_GetOrderSummary_Should_Return_One_Product_If_OrderNumber_is_2601069522001001(t *testing.T) {
	userID := 4
	orderID := 1
	trackingNumber := "KR-443947172"
	var orderNumber int64 = 2601069522001001
	updatedTime := time.Date(2026, 2, 28, 18, 58, 44, 0, time.UTC)
	expectedUpdateTime := "01-03-2026 01:58:44"

	orderDetail := order.OrderDetailWithTrackingNumber{
		ID:               orderID,
		OrderNumber:      orderNumber,
		UserID:           userID,
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		SubTotalPrice:    4314.6,
		DiscountPrice:    0,
		TotalPrice:       4364.6,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        43,
		TransactionID:    "TXN202512250934",
		Status:           "paid",
		TrackingNumber:   trackingNumber,
		Updated:          updatedTime,
	}

	orderProduct := []order.OrderProductWithPrice{
		{
			ProductBrand: "SportsFun",
			ProductName:  "Balance Training Bicycle",
			Quantity:     1,
			Price:        119.95,
		},
	}

	userDetail := auth.UserPayload{
		UserID:    userID,
		FirstName: "Noppadon",
		LastName:  "Sookwattana",
		Username:  "noppadon.s",
	}

	expected := order.OrderSummary{
		OrderNumber:    orderNumber,
		FirstName:      userDetail.FirstName,
		LastName:       userDetail.LastName,
		TrackingNumber: trackingNumber,
		ShippingMethod: "Kerry",
		PaymentMethod:  "Credit Card / Debit Card",
		OrderProductList: []order.OrderSummaryProduct{
			{
				ProductBrand:  "SportsFun",
				ProductName:   "Balance Training Bicycle",
				Quantity:      1,
				PriceTHB:      4314.6,
				TotalPriceTHB: 4314.6,
			},
		},
		SubTotalPrice:  orderDetail.SubTotalPrice,
		DiscountPrice:  orderDetail.DiscountPrice,
		TotalPrice:     orderDetail.TotalPrice,
		ShippingFee:    orderDetail.ShippingFee,
		BurnPoint:      orderDetail.BurnPoint,
		ReceivingPoint: orderDetail.EarnPoint,
		IssuedDate:     expectedUpdateTime,
	}

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderWithTrackingNumberByOrderNumber", mock.Anything, orderNumber).Return(orderDetail, nil)
	mockOrderRepository.On("GetOrderProductWithPrice", mock.Anything, orderID).Return(orderProduct, nil)

	mockUserRepository := new(mockUserRepository)
	mockUserRepository.On("FindByID", mock.Anything, userID).Return(userDetail, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		UserRepository:  mockUserRepository,
	}

	actual, err := orderService.GetOrderSummary(context.Background(), orderNumber)
	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}

func Test_ConfirmReceipt_Input_OrderNumber_Not_Found_Should_Return_ErrOrderNotFound(t *testing.T) {
	uid := 1
	var orderNumber int64 = 2601069522001099

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderByOrderNumber", mock.Anything, orderNumber).Return(order.OrderDetail{}, sql.ErrNoRows)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
	}

	err := orderService.ConfirmReceipt(context.Background(), uid, orderNumber)

	assert.Equal(t, order.ErrOrderNotFound, err)
	mockOrderRepository.AssertNotCalled(t, "UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything)
}

func Test_ConfirmReceipt_Input_OrderNumber_Owned_By_Another_User_Should_Return_ErrOrderNotOwned(t *testing.T) {
	uid := 1
	otherUserID := 2
	oid := 8004359103
	var orderNumber int64 = 2601069522001098

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderByOrderNumber", mock.Anything, orderNumber).Return(order.OrderDetail{
		ID:     oid,
		UserID: otherUserID,
	}, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
	}

	err := orderService.ConfirmReceipt(context.Background(), uid, orderNumber)

	assert.Equal(t, order.ErrOrderNotOwned, err)
	mockOrderRepository.AssertNotCalled(t, "UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything)
}

func Test_ConfirmReceipt_Input_Valid_Order_Should_Approve_Point_And_Complete_Order(t *testing.T) {
	uid := 1
	orgID := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001097

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderByOrderNumber", mock.Anything, orderNumber).Return(order.OrderDetail{
		ID:     oid,
		UserID: uid,
	}, nil)
	mockOrderRepository.On("UpdateOrderStatus", mock.Anything, oid, "completed").Return(nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("ApproveEarnPoint", mock.Anything, uid, orgID, oid).Return(nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		PointService:    mockPointInterface,
	}

	err := orderService.ConfirmReceipt(context.Background(), uid, orderNumber)

	assert.Nil(t, err)
	mockOrderRepository.AssertCalled(t, "UpdateOrderStatus", mock.Anything, oid, "completed")
}

func Test_ConfirmReceipt_Input_Order_Already_Approved_Should_Still_Succeed_Idempotently(t *testing.T) {
	uid := 1
	orgID := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001096

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderByOrderNumber", mock.Anything, orderNumber).Return(order.OrderDetail{
		ID:     oid,
		UserID: uid,
	}, nil)
	mockOrderRepository.On("UpdateOrderStatus", mock.Anything, oid, "completed").Return(nil)

	mockPointInterface := new(mockPointInterface)
	// point-service already has an APPROVED record for this order and returns it as-is (no error)
	mockPointInterface.On("ApproveEarnPoint", mock.Anything, uid, orgID, oid).Return(nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		PointService:    mockPointInterface,
	}

	err := orderService.ConfirmReceipt(context.Background(), uid, orderNumber)
	assert.Nil(t, err)

	// calling it a second time must not error either
	err = orderService.ConfirmReceipt(context.Background(), uid, orderNumber)
	assert.Nil(t, err)
}

func Test_ConfirmReceipt_Input_PointService_ErrNoPendingPoints_Should_Return_Error_Without_Completing_Order(t *testing.T) {
	uid := 1
	orgID := 1
	oid := 8004359103
	var orderNumber int64 = 2601069522001095

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderByOrderNumber", mock.Anything, orderNumber).Return(order.OrderDetail{
		ID:     oid,
		UserID: uid,
	}, nil)

	mockPointInterface := new(mockPointInterface)
	mockPointInterface.On("ApproveEarnPoint", mock.Anything, uid, orgID, oid).Return(point.ErrNoPendingPoints)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		PointService:    mockPointInterface,
	}

	err := orderService.ConfirmReceipt(context.Background(), uid, orderNumber)

	assert.Equal(t, point.ErrNoPendingPoints, err)
	mockOrderRepository.AssertNotCalled(t, "UpdateOrderStatus", mock.Anything, mock.Anything, mock.Anything)
}

func Test_GetOrderSummary_Should_Return_Two_Products_If_OrderOrderNumber_is_2601069522002002(t *testing.T) {
	userID := 5
	orderID := 2
	trackingNumber := "KR-304590466"
	var orderNumber int64 = 2601069522002002
	updatedTime := time.Date(2026, 2, 14, 1, 40, 32, 0, time.UTC)
	expectedUpdateTime := "14-02-2026 08:40:32"

	orderDetail := order.OrderDetailWithTrackingNumber{
		ID:               orderID,
		OrderNumber:      orderNumber,
		UserID:           userID,
		ShippingMethodID: 1,
		PaymentMethodID:  1,
		SubTotalPrice:    5246.22,
		DiscountPrice:    0,
		TotalPrice:       5256.22,
		ShippingFee:      50,
		BurnPoint:        0,
		EarnPoint:        52,
		TransactionID:    "TXN202512251028",
		Status:           "paid",
		TrackingNumber:   trackingNumber,
		Updated:          updatedTime,
	}

	orderProduct := []order.OrderProductWithPrice{
		{
			ProductBrand: "SportsFun",
			ProductName:  "Balance Training Bicycle",
			Quantity:     1,
			Price:        119.95,
		},
		{
			ProductBrand: "CoolKidz",
			ProductName:  "43 Piece dinner Set",
			Quantity:     2,
			Price:        12.95,
		},
	}

	userDetail := auth.UserPayload{
		UserID:    userID,
		FirstName: "Pimmida",
		LastName:  "Katethong",
		Username:  "pimmida.k",
	}

	expected := order.OrderSummary{
		OrderNumber:    orderNumber,
		FirstName:      userDetail.FirstName,
		LastName:       userDetail.LastName,
		TrackingNumber: trackingNumber,
		ShippingMethod: "Kerry",
		PaymentMethod:  "Credit Card / Debit Card",
		OrderProductList: []order.OrderSummaryProduct{
			{
				ProductBrand:  "SportsFun",
				ProductName:   "Balance Training Bicycle",
				Quantity:      1,
				PriceTHB:      4314.6,
				TotalPriceTHB: 4314.6,
			},
			{
				ProductBrand:  "CoolKidz",
				ProductName:   "43 Piece dinner Set",
				Quantity:      2,
				PriceTHB:      465.81,
				TotalPriceTHB: 931.62,
			},
		},
		SubTotalPrice:  orderDetail.SubTotalPrice,
		DiscountPrice:  orderDetail.DiscountPrice,
		TotalPrice:     orderDetail.TotalPrice,
		ShippingFee:    orderDetail.ShippingFee,
		BurnPoint:      orderDetail.BurnPoint,
		ReceivingPoint: orderDetail.EarnPoint,
		IssuedDate:     expectedUpdateTime,
	}

	mockOrderRepository := new(mockOrderRepository)
	mockOrderRepository.On("GetOrderWithTrackingNumberByOrderNumber", mock.Anything, orderNumber).Return(orderDetail, nil)
	mockOrderRepository.On("GetOrderProductWithPrice", mock.Anything, orderID).Return(orderProduct, nil)

	mockUserRepository := new(mockUserRepository)
	mockUserRepository.On("FindByID", mock.Anything, userID).Return(userDetail, nil)

	orderService := order.OrderService{
		OrderRepository: mockOrderRepository,
		UserRepository:  mockUserRepository,
	}

	actual, err := orderService.GetOrderSummary(context.Background(), orderNumber)
	assert.Equal(t, expected, actual)
	assert.Nil(t, err)
}
