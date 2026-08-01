package point

type Point struct {
	OrgID  int `json:"orgId"`
	UserID int `json:"userId"`
	Amount int `json:"amount"`
}

type TotalPoint struct {
	Point int `json:"point"`
}

type EarnPointRequest struct {
	OrgID     int     `json:"orgId"`
	UserID    int     `json:"userId"`
	OrderID   int     `json:"orderId"`
	AmountThb float64 `json:"amountThb"`
}

type ApproveEarnPointRequest struct {
	OrgID   int `json:"orgId"`
	UserID  int `json:"userId"`
	OrderID int `json:"orderId"`
}

type RedeemPointRequest struct {
	OrgID   int `json:"orgId"`
	UserID  int `json:"userId"`
	OrderID int `json:"orderId"`
	Points  int `json:"points"`
}
