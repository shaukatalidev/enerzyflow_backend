package orders

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func CreateOrder(order *Order, userID string) error {
	if order.OrderID == "" {
		return errors.New("order_id is required")
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
        INSERT INTO orders (order_id, user_id, label_id, variant, qty, cap_color, volume, created_at, updated_at, expected_delivery_date) 
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9,$10)`,
		order.OrderID,
		userID,
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

func GetOrdersByUserID(userID string, limit, offset int) ([]OrderResponse, int, error) {
	rows, err := db.DB.Query(`
        SELECT o.order_id, o.user_id, l.label_url AS label_url, 
            o.variant, o.qty, o.cap_color, o.volume, 
            o.status,o.payment_status,o.decline_reason,o.payment_screenshot_url,o.invoice_url,o.pi_url,o.created_at, o.updated_at, o.expected_delivery_date,COUNT(*) OVER() AS total_count
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.user_id = $1 
        ORDER BY o.created_at DESC 
        LIMIT $2 OFFSET $3`, userID, limit, offset)

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var (
		orders []OrderResponse
		total  int
	)
	for rows.Next() {
		var order OrderResponse
		err := rows.Scan(&order.OrderID, &order.UserID, &order.LabelURL, &order.Variant,
			&order.Qty, &order.CapColor, &order.Volume, &order.Status, &order.PaymentStatus, &order.DeclineReason, &order.PaymentUrl, &order.InvoiceUrl,&order.PiUrl, &order.CreatedAt, &order.UpdatedAt, &order.ExpectedDelivery, &total)
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func GetOrdersCountByCompanyID(userID string) (int, error) {
	var count int
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE user_id = $1`, userID).Scan(&count)
	return count, err
}

func GetOrderByID(orderID string) (*OrderResponse, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.user_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status,o.payment_status,o.decline_reason,o.payment_screenshot_url,o.invoice_url,o.pi_url,o.created_at, o.updated_at, o.expected_delivery_date
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = $1 `, orderID)

	order := &OrderResponse{}
	err := row.Scan(&order.OrderID, &order.UserID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status, &order.PaymentStatus, &order.DeclineReason, &order.PaymentUrl, &order.InvoiceUrl, &order.PiUrl,&order.CreatedAt, &order.UpdatedAt, &order.ExpectedDelivery)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func UpdateOrderStatus(orderID, status, changedBy, reason string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if status == "declined" {
		_, err = tx.Exec(`
		UPDATE orders 
		SET status = $1, decline_reason = $2, updated_at = NOW()
		WHERE order_id = $3
	`, status, reason, orderID)
	} else {
		_, err = tx.Exec(`
		UPDATE orders 
		SET status = $1, updated_at = NOW()
		WHERE order_id = $2
	`, status, orderID)
		if err != nil {
			return err
		}
	}

	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status, changed_at, changed_by, reason)
		VALUES ($1, $2, NOW(), $3, $4)
	`, orderID, status, changedBy, reason)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func UpdatePaymentStatus(orderID, paymentStatus, changedBy, reason string) error {
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
		SET payment_status = $1,
		    updated_at = NOW()
		WHERE order_id = $2
	`, paymentStatus, orderID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`
		INSERT INTO order_status_history (order_id, status, changed_at, changed_by, reason)
		VALUES ($1, $2, NOW(), $3, $4)
	`, orderID, paymentStatus, changedBy, reason)
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

func GetAllOrders(limit, offset int, role, userID string) ([]AllOrderModel, int, error) {
	baseQuery := `
	SELECT 
		o.order_id,
		o.user_id,
		c.name AS company_name,
		o.label_id,
		l.label_url,
		o.variant,
		o.qty,
		o.cap_color,
		o.volume,
		o.status,
		o.payment_status,
		o.payment_screenshot_url,
		o.invoice_url,
		o.pi_url,
		COALESCE(o.decline_reason, '') AS decline_reason,
		o.created_at,
		o.updated_at,
		u.name AS user_name,
		o.expected_delivery_date,
		COUNT(*) OVER() AS total_count
	FROM orders o
	LEFT JOIN labels l ON o.label_id = l.label_id
	INNER JOIN users u ON o.user_id = u.user_id
	INNER JOIN companies c ON u.user_id = c.user_id
	`

	var rows *sql.Rows
	var err error

	switch role {
	case "admin":
		query := baseQuery + ` ORDER BY o.created_at DESC LIMIT $1 OFFSET $2`
		rows, err = db.DB.Query(query, limit, offset)

	case "printing":
	query := `
	SELECT 
		o.order_id,
		o.user_id,
		c.name AS company_name,
		o.label_id,
		l.label_url,
		o.variant,
		o.qty,
		o.cap_color,
		o.volume,
		o.status,
		COALESCE(o.decline_reason, '') AS decline_reason,
		o.created_at,
		o.updated_at,
		u.name AS user_name,
		o.expected_delivery_date,
		COUNT(*) OVER() AS total_count
	FROM orders o
	LEFT JOIN labels l ON o.label_id = l.label_id
	INNER JOIN users u ON o.user_id = u.user_id
	INNER JOIN companies c ON u.user_id = c.user_id
	LEFT JOIN order_assignments oa ON o.order_id = oa.order_id AND oa.role = 'printing'
	WHERE 
		o.payment_status = 'payment_verified' AND
		(
			oa.user_id IS NULL OR oa.user_id = $3 OR (o.status = 'declined' AND oa.user_id = $3)
		)
		AND NOT (o.status = 'declined' AND oa.user_id IS DISTINCT FROM $3)
	ORDER BY o.created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err = db.DB.Query(query, limit, offset, userID)

case "plant":
	query := `
	SELECT 
		o.order_id,
		o.user_id,
		c.name AS company_name,
		o.label_id,
		l.label_url,
		o.variant,
		o.qty,
		o.cap_color,
		o.volume,
		o.status,
		COALESCE(o.decline_reason, '') AS decline_reason,
		o.created_at,
		o.updated_at,
		u.name AS user_name,
		o.expected_delivery_date,
		COUNT(*) OVER() AS total_count
	FROM orders o
	LEFT JOIN labels l ON o.label_id = l.label_id
	INNER JOIN users u ON o.user_id = u.user_id
	INNER JOIN companies c ON u.user_id = c.user_id
	LEFT JOIN order_assignments oa ON o.order_id = oa.order_id AND oa.role = 'plant'
	WHERE 
		o.status IN ('ready_for_plant', 'plant_processing', 'dispatched', 'completed')
		AND (oa.user_id IS NULL OR oa.user_id = $3)
	ORDER BY o.created_at DESC LIMIT $1 OFFSET $2
	`
	rows, err = db.DB.Query(query, limit, offset, userID)

	default:
		return nil, 0, fmt.Errorf("unauthorized role: %s", role)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var orders []AllOrderModel
	var total int

	for rows.Next() {
		var o AllOrderModel

		if role == "admin" {
			if err := rows.Scan(
				&o.OrderID, &o.UserID, &o.CompanyName, &o.LabelID, &o.LabelURL,
				&o.Variant, &o.Qty, &o.CapColor, &o.Volume, &o.Status,
				&o.PaymentStatus, &o.PaymentUrl, &o.InvoiceUrl, &o.PiUrl,&o.DeclineReason,
				&o.CreatedAt, &o.UpdatedAt, &o.UserName, &o.ExpectedDelivery, &total,
			); err != nil {
				return nil, 0, err
			}
		} else {
			if err := rows.Scan(
				&o.OrderID, &o.UserID, &o.CompanyName, &o.LabelID, &o.LabelURL,
				&o.Variant, &o.Qty, &o.CapColor, &o.Volume, &o.Status,
				&o.DeclineReason, &o.CreatedAt, &o.UpdatedAt, &o.UserName,
				&o.ExpectedDelivery, &total,
			); err != nil {
				return nil, 0, err
			}
		}

		orders = append(orders, o)
	}

	return orders, total, nil
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
		    payment_status = 'payment_uploaded',
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

func UpdateOrderInvoice(orderID string, urls map[string]string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	setParts := []string{}
	args := []interface{}{}
	argIdx := 1

	if invoiceURL, ok := urls["invoice_url"]; ok {
		setParts = append(setParts, "invoice_url = $" + strconv.Itoa(argIdx))
		args = append(args, invoiceURL)
		argIdx++
	}

	if piURL, ok := urls["pi_url"]; ok {
		setParts = append(setParts, "pi_url = $" + strconv.Itoa(argIdx))
		args = append(args, piURL)
		argIdx++
	}

	if len(setParts) == 0 {
		return errors.New("no URLs to update")
	}

	setParts = append(setParts, "updated_at = NOW()")

	query := "UPDATE orders SET " + strings.Join(setParts, ", ") + " WHERE order_id = $" + strconv.Itoa(argIdx)
	args = append(args, orderID)

	if _, err := tx.Exec(query, args...); err != nil {
		return fmt.Errorf("failed to update order: %w", err)
	}

	return tx.Commit()
}

func AddOrderComment(orderID, userID, role, comment string) error {
	_, err := db.DB.Exec(`
        INSERT INTO order_comments (order_id, user_id, role, comment, created_at)
        VALUES ($1, $2, $3, $4, $5)
    `, orderID, userID, role, comment, time.Now())

	return err
}

func GetCommentsByOrder(orderID string) ([]OrderComment, error) {
	rows, err := db.DB.Query(`
        SELECT id, order_id, user_id, role, comment, created_at
        FROM order_comments
        WHERE order_id = $1
        ORDER BY created_at ASC
    `, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []OrderComment
	for rows.Next() {
		var c OrderComment
		if err := rows.Scan(&c.ID, &c.OrderID, &c.UserID, &c.Role, &c.Comment, &c.CreatedAt); err != nil {
			return nil, err
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func AssignOrder(orderID, userID, role string, deadlineDays int) error {
	deadline := time.Now().Add(time.Duration(deadlineDays*24) * time.Hour)
	_, err := db.DB.Exec(`
        INSERT INTO order_assignments (order_id, user_id, role, assigned_at, deadline)
        VALUES ($1, $2, $3, NOW(), $4)
    `, orderID, userID, role, deadline)
	return err
}

func CompleteOrderAssignment(orderID, userID string) error {
	_, err := db.DB.Exec(`
        UPDATE order_assignments
        SET completed_at = $1
        WHERE order_id = $2
    `, time.Now(), orderID)
	return err
}

func IsOrderAssignedToUser(orderID, userID, role string) (bool, error) {
	var exists bool
	err := db.DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM order_assignments
			WHERE order_id = $1 AND user_id = $2 AND role = $3
		)
	`, orderID, userID, role).Scan(&exists)
	return exists, err
}