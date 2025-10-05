package orders

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
	"errors"
)

func CreateOrder(order *Order) error {
	if order.OrderID == "" || order.CompanyID == "" {
		return errors.New("order_id and company_id are required")
	}

	_, err := db.DB.Exec(`
    INSERT INTO orders (order_id, company_id, label_id, variant, qty, cap_color, volume, status) 
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		order.OrderID, order.CompanyID, order.LabelID, order.Variant, order.Qty, order.CapColor, order.Volume, order.Status)
	return err
}

func GetOrderByID(orderID string) (*Order, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status,o.payment_screenshot_url,o.invoice_url o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = $1`, orderID)

	order := &Order{}
	err := row.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status,&order.PaymentUrl,&order.InvoiceUrl, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func GetOrdersByCompanyID(companyID string, limit, offset int) ([]Order, error) {
	rows, err := db.DB.Query(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
            o.variant, o.qty, o.cap_color, o.volume, 
            o.status,o.payment_screenshot_url,o.invoice_url,o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.company_id = $1 
        ORDER BY o.created_at DESC 
        LIMIT $2 OFFSET $3`, companyID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
			&order.Qty, &order.CapColor, &order.Volume, &order.Status,&order.PaymentUrl,&order.InvoiceUrl, &order.CreatedAt, &order.UpdatedAt)
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

func GetOrderByIDAndCompanyID(orderID, companyID string) (*Order, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status,o.payment_screenshot_url,o.invoice_url,o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = $1 AND o.company_id = $2`, orderID, companyID)

	order := &Order{}
	err := row.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status,&order.PaymentUrl,&order.InvoiceUrl, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}

func UpdateOrderStatus(orderID, status, reason string) error {
	_, err := db.DB.Exec(
		`UPDATE orders SET status = $1, decline_reason = $2 WHERE order_id = $3`,
		status, reason, orderID,
	)
	return err
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
		u.name AS user_name
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
			&o.CapColor, &o.Volume, &o.Status,&o.PaymentUrl,&o.InvoiceUrl,&o.DeclineReason, &o.CreatedAt, &o.UpdatedAt, &o.UserName); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	return orders, nil
}

func UpdateOrderPaymentScreenshot(orderID, screenshotURL string) error {
	if orderID == "" || screenshotURL == "" {
		return errors.New("orderID and screenshotURL cannot be empty")
	}

	_, err := db.DB.Exec(
		`UPDATE orders 
		 SET payment_screenshot_url = $1, 
		     status = 'payment_uploaded', 
		     updated_at = NOW()
		 WHERE order_id = $2`,
		screenshotURL, orderID,
	)
	return err
}
