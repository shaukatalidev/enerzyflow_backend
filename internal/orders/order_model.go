package orders

import "time"

type Order struct {
	OrderID          string    `json:"order_id"`
	UserID           string    `json:"user_id"`
	LabelID          string    `json:"label_id"`
	LabelURL         string    `json:"label_url"`
	Variant          string    `json:"variant"`
	Qty              int       `json:"qty"`
	CapColor         string    `json:"cap_color"`
	Volume           int       `json:"volume"`
	Status           string    `json:"status"`
	PaymentStatus    string    `json:"payment_status"`
	PaymentUrl       string    `json:"payment_url"`
	InvoiceUrl       string    `json:"invoice_url"`
	PiUrl			string    `json:"pi_url"`
	ExpectedDelivery time.Time `json:"expected_delivery"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateOrderRequest struct {
	LabelID  string `json:"label_id" binding:"required"`
	Variant  string `json:"variant" binding:"required"`
	Qty      int    `json:"qty" binding:"required,min=1"`
	CapColor string `json:"cap_color" binding:"required"`
	Volume   int    `json:"volume" binding:"required,min=1"`
}

type OrderResponse struct {
	OrderID          string    `json:"order_id"`
	UserID           string    `json:"user_id"`
	LabelURL         string    `json:"label_url"`
	Variant          string    `json:"variant"`
	Qty              int       `json:"qty"`
	CapColor         string    `json:"cap_color"`
	Volume           int       `json:"volume"`
	Status           string    `json:"status"`
	PaymentStatus    string    `json:"payment_status"`
	DeclineReason    string    `json:"decline_reason"`
	PaymentUrl       string    `json:"payment_url"`
	InvoiceUrl       string    `json:"invoice_url"`
	PiUrl			string    `json:"pi_url"`
	ExpectedDelivery time.Time `json:"expected_delivery"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type OrderListResponse struct {
	Orders []OrderResponse `json:"orders"`
	Total  int             `json:"total"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

type AllOrderModel struct {
	OrderID          string    `json:"order_id" db:"order_id"`
	UserID           string    `json:"user_id"`
	UserName         string    `json:"user_name" db:"user_name"`
	CompanyName      string    `json:"company_name" db:"company_name"`
	LabelID          string    `json:"label_id" db:"label_id"`
	LabelURL         string    `json:"label_url" db:"label_url"`
	Variant          string    `json:"variant" db:"variant"`
	Qty              int       `json:"qty" db:"qty"`
	CapColor         string    `json:"cap_color" db:"cap_color"`
	Volume           string    `json:"volume" db:"volume"`
	Status           string    `json:"status,omitempty" db:"status"`
	PaymentStatus    string    `json:"payment_status,omitempty"`
	DeclineReason    string    `json:"decline_reason"`
	PaymentUrl       string    `json:"payment_url,omitempty"`
	InvoiceUrl       string    `json:"invoice_url"`
	PiUrl            string    `json:"pi_url"`
	ExpectedDelivery time.Time `json:"expected_delivery" db:"expected_delivery_date"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

type OrderStatusHistory struct {
	Status    string    `json:"status"`
	ChangedAt time.Time `json:"changed_at"`
	ChangedBy string    `json:"changed_by,omitempty"`
	Reason    string    `json:"reason,omitempty"`
}

type OrderComment struct {
	ID        int       `json:"id"`
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Role      string    `json:"role"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderLabelDetails struct {
    ID             int       `json:"id"`
    OrderID        string    `json:"order_id"`
    NoOfSheets     int       `json:"no_of_sheets"`
    CuttingType    string    `json:"cutting_type"`
    LabelsPerSheet int       `json:"labels_per_sheet"`
    Description    string    `json:"description"`
}

type SaveLabelDetailsRequest struct {
    NoOfSheets     int    `json:"no_of_sheets" binding:"required"`
    CuttingType    string `json:"cutting_type" binding:"required"`
    LabelsPerSheet int    `json:"labels_per_sheet" binding:"required"`
    Description    string `json:"description"`
}
