package point

type SubmitedPoint struct {
	Amount  int `json:"amount"`
	OrderID int `json:"orderId"`
}

type Point struct {
	OrgID   int `json:"orgId"`
	UserID  int `json:"userId"`
	OrderID int `json:"orderId"`
	Amount  int `json:"amount"`
}

type TotalPoint struct {
	Point int `json:"point"`
}

type BalanceItem struct {
	Status string `json:"status"`
	Point  int    `json:"point"`
}

type ApproveEarnPointRequest struct {
	OrgID   int `json:"orgId"`
	UserID  int `json:"userId"`
	OrderID int `json:"orderId"`
}
