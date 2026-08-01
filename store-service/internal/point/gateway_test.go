package point_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"store-service/internal/point"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetBalance_Input_Gateway_Response_200_Should_Be_Point_From_Body(t *testing.T) {
	var requestURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"point":150}`))
	}))
	defer server.Close()

	gateway := point.PointGateway{
		PointEndpoint: server.URL,
	}
	actual, err := gateway.GetBalance(context.Background(), 1, 1)

	assert.Equal(t, 150, actual)
	assert.Equal(t, nil, err)
	assert.Equal(t, "/api/v1/point/balance?orgId=1&userId=1", requestURL)
}

func Test_GetBalance_Input_Gateway_Response_500_Should_Be_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	gateway := point.PointGateway{
		PointEndpoint: server.URL,
	}
	actual, err := gateway.GetBalance(context.Background(), 1, 1)

	assert.Equal(t, 0, actual)
	assert.EqualError(t, err, "response is not ok but it's 500")
}

func Test_ApproveEarnPoint_Input_Gateway_Response_200_Should_Be_No_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	gateway := point.PointGateway{
		PointEndpoint: server.URL,
	}
	err := gateway.ApproveEarnPoint(context.Background(), point.ApproveEarnPointRequest{
		OrgID:   1,
		UserID:  1,
		OrderID: 10,
	})

	assert.Equal(t, nil, err)
}

func Test_ApproveEarnPoint_Input_Gateway_Response_404_Should_Be_ErrNoPendingPoints(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	gateway := point.PointGateway{
		PointEndpoint: server.URL,
	}
	err := gateway.ApproveEarnPoint(context.Background(), point.ApproveEarnPointRequest{
		OrgID:   1,
		UserID:  1,
		OrderID: 10,
	})

	assert.ErrorIs(t, err, point.ErrNoPendingPoints)
}

func Test_ApproveEarnPoint_Input_Gateway_Response_500_Should_Be_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	gateway := point.PointGateway{
		PointEndpoint: server.URL,
	}
	err := gateway.ApproveEarnPoint(context.Background(), point.ApproveEarnPointRequest{
		OrgID:   1,
		UserID:  1,
		OrderID: 10,
	})

	assert.EqualError(t, err, "response is not ok but it's 500")
}
