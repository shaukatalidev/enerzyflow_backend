package orders

import (
	"context"
	"enerzyflow_backend/internal/companies"
	"enerzyflow_backend/utils"
	"errors"
	"fmt"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
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
		OrderID:          uuid.New().String(),
		UserID:           userID,
		LabelID:          req.LabelID,
		Variant:          req.Variant,
		Qty:              req.Qty,
		CapColor:         req.CapColor,
		Volume:           req.Volume,
		Status:           "placed",
		ExpectedDelivery: time.Now().Add(10 * 24 * time.Hour),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := CreateOrder(order, userID); err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	return &OrderResponse{
		OrderID:          order.OrderID,
		UserID:           userID,
		LabelURL:         label.URL,
		Variant:          req.Variant,
		Qty:              req.Qty,
		CapColor:         req.CapColor,
		Volume:           req.Volume,
		Status:           "placed",
		PaymentStatus:    "payment_pending",
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
	if order.UserID != userID {
		return nil, errors.New("unauthorized access to order")
	}

	return &OrderResponse{
		OrderID:          order.OrderID,
		UserID:           order.UserID,
		LabelURL:         order.LabelURL,
		Variant:          order.Variant,
		Qty:              order.Qty,
		CapColor:         order.CapColor,
		Volume:           order.Volume,
		Status:           order.Status,
		PaymentStatus:    order.PaymentStatus,
		DeclineReason:    order.DeclineReason,
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

	orders, total, err := GetOrdersByUserID(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	orderResponses := make([]OrderResponse, len(orders))
	for i, order := range orders {
		orderResponses[i] = OrderResponse{
			OrderID:          order.OrderID,
			UserID:           userID,
			LabelURL:         order.LabelURL,
			Variant:          order.Variant,
			Qty:              order.Qty,
			CapColor:         order.CapColor,
			Volume:           order.Volume,
			Status:           order.Status,
			PaymentStatus:    order.PaymentStatus,
			DeclineReason:    order.DeclineReason,
			PaymentUrl:       order.PaymentUrl,
			InvoiceUrl:       order.InvoiceUrl,
			PiUrl:            order.PiUrl,
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

func GetAllOrdersService(role string, limit, offset int, userID string) ([]AllOrderModel, int, error) {
	return GetAllOrders(limit, offset, role, userID)
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
		validStatuses := map[string]bool{
			"placed":           true,
			"printing":         true,
			"ready_for_plant":  true,
			"plant_processing": true,
			"dispatched":       true,
			"completed":        true,
			"declined":         true,
		}

		if !validStatuses[req.Status] {
			return fmt.Errorf("invalid status '%s' for admin", req.Status)
		}
		if req.Status == "declined" {
			reason := strings.TrimSpace(req.Reason)
			reason = strings.Trim(reason, `"`)
			if reason == "" {
				return errors.New("reason required when declining an order")
			}
			req.Reason = reason
		}

		if req.Status != "declined" && order.PaymentStatus != "payment_verified" {
			return fmt.Errorf("cannot update order status until payment is verified")
		}

		return UpdateOrderStatus(orderID, req.Status, userID, req.Reason)

	case "printing":
		if order.PaymentStatus != "payment_verified" {
			return errors.New("printing can only handle payment-verified orders")
		}

		switch order.Status {
		case "placed":
			switch req.Status {
			case "accepted":
				if err := AssignOrder(orderID, userID, "printing", 2); err != nil {
					return err
				}

				return UpdateOrderStatus(orderID, "printing", userID, "")
			case "declined":
				if req.Reason == "" {
					return errors.New("reason required when declining order")
				}
				return UpdateOrderStatus(orderID, "declined", userID, req.Reason)
			default:
				return errors.New("printing can only accept or decline orders")
			}
		case "printing":
			if req.Status == "ready_for_plant" {
				if err := UpdateOrderStatus(orderID, "ready_for_plant", userID, ""); err != nil {
					return err
				}

				if err := CompleteOrderAssignment(orderID, userID); err != nil {
					return err
				}
				return nil
			}

			return errors.New("invalid status update from printing")
		default:
			return errors.New("printing cannot handle this status")
		}

	case "plant":
		switch order.Status {
		case "ready_for_plant":
			if err := AssignOrder(orderID, userID, "plant", 3); err != nil {
				return err
			}
			return UpdateOrderStatus(orderID, "plant_processing", userID, "")
		case "plant_processing":
			if err := CompleteOrderAssignment(orderID, userID); err != nil {
				return err
			}
			return UpdateOrderStatus(orderID, "dispatched", userID, "")
		default:
			return errors.New("plant can only handle 'ready_for_plant' or 'plant_processing' statuses")
		}

	default:
		return errors.New("unauthorized role")
	}
}

func UpdatePaymentStatusService(orderID, paymentStatus, reason, adminID string) error {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}

	if order.PaymentStatus != "payment_uploaded" {
		return errors.New("cannot update payment: payment not uploaded yet")
	}

	switch paymentStatus {
	case "payment_verified":
		return UpdatePaymentStatus(orderID, "payment_verified", adminID, "")

	case "payment_rejected":
		if reason == "" {
			return errors.New("reason required when rejecting payment")
		}
		return UpdatePaymentStatus(orderID, "payment_rejected", adminID, reason)

	default:
		return errors.New("invalid payment status")
	}
}

func GetOrderTrackingService(orderID, userID, role string) ([]OrderStatusHistory, error) {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if role == "business_owner" {
		if order.UserID != userID {
			return nil, errors.New("unauthorized access to order")
		}
	}

	return GetOrderStatusHistory(orderID)
}

func UploadPaymentScreenshotService(orderID string, fileHeader *multipart.FileHeader, userID string) (string, error) {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return "", err
	}
	if order == nil {
		return "", errors.New("order not found")
	}
	if order.UserID != userID {
		return "", errors.New("unauthorized")
	}

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

func UploadInvoiceService(orderID string, invoiceFile,piFile *multipart.FileHeader) (map[string]string, error) {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	resutl := map[string]string{}

	if invoiceFile != nil {
		invoiceURL, err := utils.UploadFileToCloud(invoiceFile, "invoices", "invoice_"+orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload invoice: %w", err)
		}
		resutl["invoice_url"] = invoiceURL
	}
	if piFile != nil {
		piURL, err := utils.UploadFileToCloud(piFile, "pi", "pi_"+orderID)
		if err != nil {
			return nil, fmt.Errorf("failed to upload pi: %w", err)
		}
		resutl["pi_url"] = piURL
	}

	if err := UpdateOrderInvoice(orderID, resutl); err != nil {
			return nil, fmt.Errorf("failed to update order: %w", err)
	}

	return resutl, nil
}

func AddOrderCommentService(orderID, userID, role, comment string) error {
	order, err := GetOrderByID(orderID)
	if err != nil {
		return err
	}
	if order == nil {
		return errors.New("order not found")
	}
	if comment == "" {
		return errors.New("comment cannot be empty")
	}

	assigned, err := IsOrderAssignedToUser(orderID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to verify assignment: %v", err)
	}
	if !assigned {
		return errors.New("you are not assigned to this order")
	}

	switch role {
	case "printing":
		if order.Status != "printing" {
			return errors.New("printing can only comment on orders in 'printing' status")
		}
	case "plant":
		if order.Status != "plant_processing" {
			return errors.New("plant can only comment on orders in 'plant_processing' status")
		}
	default:
		return errors.New("unauthorized role")
	}

	return AddOrderComment(orderID, userID, role, comment)
}

func GetOrderCommentsService(orderID string, role string) ([]OrderComment, error) {
	if role != "admin" {
		return nil, errors.New("unauthorized: only admin can view comments")
	}
	return GetCommentsByOrder(orderID)
}

func SaveOrderLabelDetailsService(orderID string, noOfSheets int, cuttingType string, labelsPerSheet int, description string) error {
    order, err := GetOrderByID(orderID)
    if err != nil {
        return err
    }
    if order == nil {
        return errors.New("order not found")
    }

    details := OrderLabelDetails{
        OrderID:        orderID,
        NoOfSheets:     noOfSheets,
        CuttingType:    cuttingType,
        LabelsPerSheet: labelsPerSheet,
        Description:    description,
    }

    return SaveOrderLabelDetails(details)
}

func GetOrderLabelDetailsService(orderID,userID,role string) (*OrderLabelDetails, error) {
	switch role {
	case "admin":
		return GetOrderLabelDetails(orderID)

	case "printing":
		assigned, err := IsOrderAssignedToUser(orderID, userID, role)
		if err != nil {
			return nil, fmt.Errorf("failed to verify assignment: %v", err)
		}
		if !assigned {
			return nil, errors.New("you are not assigned to this order")
		}
		return GetOrderLabelDetails(orderID)

	default:
		return nil, errors.New("unauthorized role")
	}
    
}