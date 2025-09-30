package orders

import "time"

type Order struct {
	OrderID   string    `json:"order_id"`
	CompanyID string    `json:"company_id"`
	LabelID   string    `json:"label_id"` 
	LabelURL   string    `json:"label_url"`
	Variant   string    `json:"variant"`
	Qty       int       `json:"qty"`
	CapColor  string    `json:"cap_color"`
	Volume    int       `json:"volume"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	LabelID  string `json:"label_id" binding:"required"`
	Variant  string `json:"variant" binding:"required"`
	Qty      int    `json:"qty" binding:"required,min=1"`
	CapColor string `json:"cap_color" binding:"required"`
	Volume   int    `json:"volume" binding:"required,min=1"`
}

type OrderResponse struct {
	OrderID   string    `json:"order_id"`
	CompanyID string    `json:"company_id"`
	LabelURL  string    `json:"label_url"`
	Variant   string    `json:"variant"`
	Qty       int       `json:"qty"`
	CapColor  string    `json:"cap_color"`
	Volume    int       `json:"volume"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
}