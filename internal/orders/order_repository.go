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
        VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		order.OrderID, order.CompanyID, order.LabelURL, order.Variant, order.Qty, order.CapColor, order.Volume, order.Status)
	return err
}

func GetOrderByID(orderID string) (*Order, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status, o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = ?`, orderID)

	order := &Order{}
	err := row.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status, &order.CreatedAt, &order.UpdatedAt)

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
               o.status, o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.company_id = ? 
        ORDER BY o.created_at DESC 
        LIMIT ? OFFSET ?`, companyID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
			&order.Qty, &order.CapColor, &order.Volume, &order.Status, &order.CreatedAt, &order.UpdatedAt)
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
	err := db.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE company_id = ?`, companyID).Scan(&count)
	return count, err
}


func GetOrderByIDAndCompanyID(orderID, companyID string) (*Order, error) {
	row := db.DB.QueryRow(`
        SELECT o.order_id, o.company_id, l.label_url AS label_url, 
               o.variant, o.qty, o.cap_color, o.volume, 
               o.status, o.created_at, o.updated_at
        FROM orders o
        LEFT JOIN labels l ON o.label_id = l.label_id
        WHERE o.order_id = ? AND o.company_id = ?`, orderID, companyID)

	order := &Order{}
	err := row.Scan(&order.OrderID, &order.CompanyID, &order.LabelURL, &order.Variant,
		&order.Qty, &order.CapColor, &order.Volume, &order.Status, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return order, nil
}
