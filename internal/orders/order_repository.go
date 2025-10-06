package orders

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
	"errors"
	"fmt"
)

func CreateOrder(order *Order, userID string) error {
	if order.OrderID == "" || order.CompanyID == "" {
		return errors.New("order_id and company_id are required")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
        INSERT INTO orders (order_id, company_id, label_id, variant, qty, cap_color, volume, created_at, updated_at, expected_delivery_date) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10)`,
		order.OrderID,
		order.CompanyID,
		order.LabelID,
		order.Variant,
		order.Qty,
		order.CapColor,
		order.Volume,
		order.CreatedAt,
		order.UpdatedAt,
		order.ExpectedDelivery,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = tx.Exec(`
        INSERT INTO order_status_history (order_id, status, changed_at, changed_by)
        VALUES ($1, $2, NOW(), $3)
    `, order.OrderID, order.Status, userID)
	if err != nil {
		return fmt.Errorf("failed to insert initial status history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func GetOrdersByCompanyID(companyID string, limit, offset int) ([]OrderResponse, error) {
	rows, err := db.DB.Query(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
            o.variant, o.qty, o.cap_color, o.volume, 
            o.status,o.decline_reason,o.payment_screenshot_url,o.invoice_url,o.created_at, o.updated_at, o.expected_delivery_date
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.company_id = $1 
        ORDER BY o.created_at DESC 
        LIMIT $2 OFFSET $3`, companyID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderResponse
	for rows.Next() {
		var order OrderResponse
		err := rows.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
			&order.Qty, &order.CapColor, &order.Volume, &order.Status,&order.DeclineReason,&order.PaymentUrl,&order.InvoiceUrl, &order.CreatedAt, &order.UpdatedAt,&order.ExpectedDelivery)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func GetOrdersCountByCompanyID(companyID string) (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE company_id = $1`, companyID).Scan(&count)
	return count, err
}

func GetOrderByIDAndCompanyID(orderID, companyID string) (*OrderResponse, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status,o.decline_reason,o.payment_screenshot_url,o.invoice_url,o.created_at, o.updated_at, o.expected_delivery_date
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = $1 AND o.company_id = $2`, orderID, companyID)

	order := &OrderResponse{}
	err := row.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status,&order.DeclineReason,&order.PaymentUrl,&order.InvoiceUrl, &order.CreatedAt, &order.UpdatedAt,&order.ExpectedDelivery)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func UpdateOrderStatus(orderID, status, reason, changedBy string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1️⃣ Update current status in orders table
	_, err = tx.Exec(`
		UPDATE orders 
		SET status = $1, updated_at = NOW() 
		WHERE order_id = $2
	`, status, orderID)
	if err != nil {
		return err
	}

	// 2️⃣ Insert into status history
	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status, changed_at, changed_by, reason)
		VALUES ($1, $2, NOW(), $3, $4)
	`, orderID, status, changedBy, reason)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func GetOrderStatusHistory(orderID string) ([]OrderStatusHistory, error) {
	rows, err := db.DB.Query(`
		SELECT status, changed_at, changed_by, reason
		FROM order_status_history
		WHERE order_id = $1
		ORDER BY changed_at ASC
	`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []OrderStatusHistory
	for rows.Next() {
		var h OrderStatusHistory
		if err := rows.Scan(&h.Status, &h.ChangedAt, &h.ChangedBy, &h.Reason); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, nil
}

func GetAllOrders(limit, offset int) ([]AllOrderModel, error) {
	query := `
	SELECT 
		o.order_id,
		o.company_id,
		c.name AS company_name,
		o.label_id,
		l.label_url,
		o.variant,
		o.qty,
		o.cap_color,
		o.volume,
		o.status,
		o.payment_screenshot_url,
		o.invoice_url,
		COALESCE(o.decline_reason, '') AS decline_reason,
		o.created_at,
		o.updated_at,
		u.name AS user_name,
		o.expected_delivery_date
	FROM orders o
	LEFT JOIN labels l ON o.label_id = l.label_id
	INNER JOIN companies c ON o.company_id = c.company_id
	INNER JOIN users u ON c.user_id = u.user_id
	WHERE o.status = 'placed'
	ORDER BY o.created_at DESC
	LIMIT $1 OFFSET $2
`
	rows, err := db.DB.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []AllOrderModel
	for rows.Next() {
		var o AllOrderModel
		if err := rows.Scan(
			&o.OrderID, &o.CompanyID, &o.CompanyName, &o.LabelID, &o.LabelURL, &o.Variant, &o.Qty,
			&o.CapColor, &o.Volume, &o.Status,&o.PaymentUrl,&o.InvoiceUrl,&o.DeclineReason, &o.CreatedAt, &o.UpdatedAt, &o.UserName,&o.ExpectedDelivery); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func UpdateOrderPaymentScreenshot(orderID, screenshotURL, userID string) error {
	if orderID == "" || screenshotURL == "" {
		return errors.New("orderID and screenshotURL cannot be empty")
	}

	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	_, err = tx.Exec(`
		UPDATE orders
		SET payment_screenshot_url = $1,
		    status = 'payment_uploaded',
		    updated_at = NOW()
		WHERE order_id = $2
	`, screenshotURL, orderID)
	if err != nil {
		return fmt.Errorf("failed to update order payment screenshot: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status, changed_at, changed_by)
		VALUES ($1, $2, NOW(), $3)
	`, orderID, "payment_uploaded", userID)
	if err != nil {
		return fmt.Errorf("failed to insert into order_status_history: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}