package point

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type PointGateway struct {
	PointEndpoint string
}

func (gateway PointGateway) GetPoints(ctx context.Context, uid int) ([]Point, error) {
	endPoint := gateway.PointEndpoint + "/api/v1/point"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endPoint, nil)
	if err != nil {
		return []Point{}, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return []Point{}, err
	}
	if response.StatusCode != 200 {
		return []Point{}, fmt.Errorf("response is not ok but it's %d", response.StatusCode)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return []Point{}, err
	}

	var PointGatewayResponse []Point
	err = json.Unmarshal(responseData, &PointGatewayResponse)
	if err != nil {
		return []Point{}, err
	}

	return PointGatewayResponse, nil
}

func (gateway PointGateway) GetBalance(ctx context.Context, orgID int, uid int) (int, error) {
	endPoint := fmt.Sprintf("%s/api/v1/point/balance?orgId=%d&userId=%d", gateway.PointEndpoint, orgID, uid)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endPoint, nil)
	if err != nil {
		return 0, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	if response.StatusCode != 200 {
		return 0, fmt.Errorf("response is not ok but it's %d", response.StatusCode)
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}

	var balanceResponse TotalPoint
	err = json.Unmarshal(responseData, &balanceResponse)
	if err != nil {
		return 0, err
	}

	return balanceResponse.Point, nil
}

func (gateway PointGateway) CreateEarnPoint(ctx context.Context, body EarnPointRequest) error {
	data, _ := json.Marshal(body)
	endPoint := gateway.PointEndpoint + "/api/v1/point"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endPoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("response is not ok but it's %d", response.StatusCode)
	}
	return nil
}

func (gateway PointGateway) ApproveEarnPoint(ctx context.Context, body ApproveEarnPointRequest) error {
	data, _ := json.Marshal(body)
	endPoint := gateway.PointEndpoint + "/api/v1/point/approve"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endPoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return ErrNoPendingPoints
	}
	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("response is not ok but it's %d", response.StatusCode)
	}
	return nil
}

func (gateway PointGateway) RedeemPoint(ctx context.Context, body RedeemPointRequest) error {
	data, _ := json.Marshal(body)
	endPoint := gateway.PointEndpoint + "/api/v1/point/redeem"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endPoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusBadRequest {
		return ErrInsufficientPoints
	}
	if response.StatusCode != 200 && response.StatusCode != 201 {
		return fmt.Errorf("response is not ok but it's %d", response.StatusCode)
	}
	return nil
}
