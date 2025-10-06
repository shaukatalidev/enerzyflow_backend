package orders

import (
	"context"
	"enerzyflow_backend/internal/companies"
	"errors"
	"fmt"
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/google/uuid"
	"mime/multipart"
	"os"
	"time"
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
		OrderID:          uuid.New().String(),
		CompanyID:        company.CompanyID,
		LabelID:          req.LabelID,
		Variant:          req.Variant,
		Qty:              req.Qty,
		CapColor:         req.CapColor,
		Volume:           req.Volume,
		Status:           "payment_pending",
		ExpectedDelivery: time.Now().Add(5 * 24 * time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := CreateOrder(order, userID); err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	return &OrderResponse{
		OrderID:          order.OrderID,
		CompanyID:        company.CompanyID,
		LabelURL:         label.URL,
		Variant:          req.Variant,
		Qty:              req.Qty,
		CapColor:         req.CapColor,
		Volume:           req.Volume,
		Status:           "payment_pending",
		ExpectedDelivery: order.ExpectedDelivery,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

func GetOrderService(userID, orderID string) (*OrderResponse, error) {
	if userID == "" {
		return nil, errors.New("missing authenticated user id")
	}

	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	return &OrderResponse{
		OrderID:          order.OrderID,
		CompanyID:        order.CompanyID,
		LabelURL:         order.LabelURL,
		Variant:          order.Variant,
		Qty:              order.Qty,
		CapColor:         order.CapColor,
		Volume:           order.Volume,
		Status:           order.Status,
		PaymentUrl:       order.PaymentUrl,
		InvoiceUrl:       order.InvoiceUrl,
		ExpectedDelivery: order.ExpectedDelivery,
		CreatedAt:        order.CreatedAt,
		UpdatedAt:        order.UpdatedAt,
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
			OrderID:          order.OrderID,
			CompanyID:        order.CompanyID,
			LabelURL:         order.LabelURL,
			Variant:          order.Variant,
			Qty:              order.Qty,
			CapColor:         order.CapColor,
			Volume:           order.Volume,
			Status:           order.Status,
			DeclineReason:    order.DeclineReason,
			PaymentUrl:       order.PaymentUrl,
			InvoiceUrl:       order.InvoiceUrl,
			ExpectedDelivery: order.ExpectedDelivery,
			CreatedAt:        order.CreatedAt,
			UpdatedAt:        order.UpdatedAt,
		}
	}

	return &OrderListResponse{
		Orders: orderResponses,
		Total:  total,
	}, nil
}

func GetAllOrdersService(role string, limit, offset int) ([]AllOrderModel, error) {
	switch role {
	case "admin":
		return GetAllOrders(limit, offset, nil)
	case "printing":
		status := "payment_verified"
		return GetAllOrders(limit, offset, &status)
	default:
		return nil, fmt.Errorf("unauthorized role: %s", role)
	}
}

func UpdateOrderStatusService(userID, role, orderID string, req UpdateOrderStatusRequest) error {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	switch role {
	case "admin":
		if req.Status == "declined" || req.Status == "payment_rejected" {
			if req.Reason == "" {
				return errors.New("reason is required for declined orders")
			}
		}
	case "printing":
		if order.Status != "placed" {
			return errors.New("printing role can only update orders with status 'placed'")
		}
		if req.Status != "accepted" && req.Status != "declined" {
			return errors.New("printing role can only accept or decline an order")
		}
		if req.Status == "declined" && req.Reason == "" {
			return errors.New("reason is required when declining order")
		}
		if req.Status == "accepted" {
			req.Status = "printing"
			req.Reason = ""
		}
	default:
		return errors.New("unauthorized role")
	}
	return UpdateOrderStatus(orderID, req.Status, userID, req.Reason)
}

func GetOrderTrackingService(orderID string) ([]OrderStatusHistory, error) {
	return GetOrderStatusHistory(orderID)
}

func UploadPaymentScreenshotService(orderID string, fileHeader *multipart.FileHeader, userID string) (string, error) {
	if fileHeader == nil {
		return "", errors.New("file cannot be nil")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	uploadResult, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
		Folder:   "orders/payment_screenshots",
		PublicID: orderID,
	})
	if err != nil {
		return "", err
	}

	if err := UpdateOrderPaymentScreenshot(orderID, uploadResult.SecureURL, userID); err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

func UploadInvoiceService(orderID string, file *multipart.FileHeader) (string, error) {
	if file == nil {
		return "", errors.New("no file provided")
	}

	cld, err := cloudinary.NewFromParams(
		os.Getenv("CLOUDINARY_CLOUD_NAME"),
		os.Getenv("CLOUDINARY_API_KEY"),
		os.Getenv("CLOUDINARY_API_SECRET"),
	)
	if err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	uploadResult, err := cld.Upload.Upload(context.Background(), src, uploader.UploadParams{
		Folder: "enerzyflow/invoices",
		ResourceType: "raw", 
		PublicID:     "invoice_" + orderID,
	})
	if err != nil {
		return "", err
	}
	
	if err := UpdateOrderInvoice(orderID, uploadResult.SecureURL); err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}