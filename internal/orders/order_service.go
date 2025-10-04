package orders

import (
	"enerzyflow_backend/internal/companies"
	"errors"
	"fmt"
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

	order := &Order{
		OrderID:   uuid.New().String(),
		CompanyID: company.CompanyID,
		LabelID:   req.LabelID,
		Variant:   req.Variant,
		Qty:       req.Qty,
		CapColor:  req.CapColor,
		Volume:    req.Volume,
		Status:    "placed",
	}

	if err := CreateOrder(order); err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}


	return &OrderResponse{
		OrderID:   order.OrderID,
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

func GetAllOrdersService(userID string, limit, offset int) ([]AllOrderModel, error) {
	return GetAllOrders(limit, offset)
}

func UpdateOrderStatusService(userID, orderID string, req UpdateOrderStatusRequest) error {
	fmt.Println(req.Status)
	if req.Status == "declined" && req.Reason == "" {
		return errors.New("reason is required when canceling order")
	}

	newStatus := req.Status
	if req.Status == "accepted" {
		newStatus = "printing"
		req.Reason = ""
	}

	return UpdateOrderStatus(orderID, newStatus, req.Reason)
}