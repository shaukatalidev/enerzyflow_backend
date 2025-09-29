package orders

import (
	"enerzyflow_backend/internal/companies"
	"enerzyflow_backend/internal/db"
	"errors"
	"time"

	"github.com/google/uuid"
)

func CreateOrderService(userID string, req CreateOrderRequest) (*OrderResponse, error) {
	if userID == "" {
		return nil, errors.New("missing authenticated user id")
	}

	company, err := companies.GetCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New("company not found for user")
	}

	label, err := companies.GetLabelByIDAndCompanyID(req.LabelID, company.CompanyID)
	if err != nil {
		return nil, errors.New("failed to validate label: " + err.Error())
	}
	if label == nil {
		return nil, errors.New("label does not belong to your company")
	}

	orderID := uuid.New().String()
	_, err = db.DB.Exec(`
		INSERT INTO orders (order_id, company_id, label_id, variant, qty, cap_color, volume, status) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		orderID, company.CompanyID, req.LabelID, req.Variant, req.Qty, req.CapColor, req.Volume, "placed")
	if err != nil {
		return nil, err
	}

	return &OrderResponse{
		OrderID:   orderID,
		CompanyID: company.CompanyID,
		LabelURL:  label.URL,
		Variant:   req.Variant,
		Qty:       req.Qty,
		CapColor:  req.CapColor,
		Volume:    req.Volume,
		Status:    "placed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil
}

func GetOrderService(userID, orderID string) (*OrderResponse, error) {
	if userID == "" {
		return nil, errors.New("missing authenticated user id")
	}

	company, err := companies.GetCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New("company not found for user")
	}

	order, err := GetOrderByIDAndCompanyID(orderID, company.CompanyID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	return &OrderResponse{
		OrderID:   order.OrderID,
		CompanyID: order.CompanyID,
		LabelURL:  order.LabelURL,
		Variant:   order.Variant,
		Qty:       order.Qty,
		CapColor:  order.CapColor,
		Volume:    order.Volume,
		Status:    order.Status,
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}, nil
}

func GetOrdersService(userID string, limit, offset int) (*OrderListResponse, error) {
	if userID == "" {
		return nil, errors.New("missing authenticated user id")
	}

	company, err := companies.GetCompanyByUserID(userID)
	if err != nil {
		return nil, err
	}
	if company == nil {
		return nil, errors.New("company not found for user")
	}

	orders, err := GetOrdersByCompanyID(company.CompanyID, limit, offset)
	if err != nil {
		return nil, err
	}

	total, err := GetOrdersCountByCompanyID(company.CompanyID)
	if err != nil {
		return nil, err
	}

	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = OrderResponse{
			OrderID:   order.OrderID,
			CompanyID: order.CompanyID,
			LabelURL:  order.LabelURL,
			Variant:   order.Variant,
			Qty:       order.Qty,
			CapColor:  order.CapColor,
			Volume:    order.Volume,
			Status:    order.Status,
			CreatedAt: order.CreatedAt,
			UpdatedAt: order.UpdatedAt,
		}
	}

	return &OrderListResponse{
		Orders: orderResponses,
		Total:  total,
	}, nil
}

