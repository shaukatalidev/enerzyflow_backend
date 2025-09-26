package companies

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
)

func UpsertCompanyTx(tx *sql.Tx, c *Company) error {
    // Try update first
    res, err := tx.Exec(`UPDATE companies SET name = ?, address = ?, logo = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, c.Name, c.Address, c.Logo, c.UserID)
    if err != nil {
        return err
    }
    rows, err := res.RowsAffected()
    if err != nil {
        return err
    }
    if rows == 0 {
        _, err = tx.Exec(`INSERT INTO companies (company_id, user_id, name, address, logo) VALUES (?, ?, ?, ?, ?)`, c.CompanyID, c.UserID, c.Name, c.Address, c.Logo)
        return err
    }
    return nil
}

func ReplaceCompanyOutletsTx(tx *sql.Tx, companyID string, outlets []CompanyOutlet) error {
    if _, err := tx.Exec(`DELETE FROM company_outlets WHERE company_id = ?`, companyID); err != nil {
        return err
    }
    for _, o := range outlets {
        if _, err := tx.Exec(`INSERT INTO company_outlets (id, company_id, name, address) VALUES (?, ?, ?, ?)`, o.ID, o.CompanyID, o.Name, o.Address); err != nil {
            return err
        }
    }
    return nil
}

// Non-transactional helpers
func GetCompanyByUserID(userID string) (*Company, error) {
    row := db.DB.QueryRow(`SELECT company_id, user_id, name, address, logo FROM companies WHERE user_id = ?`, userID)
    c := &Company{}
    if err := row.Scan(&c.CompanyID, &c.UserID, &c.Name, &c.Address, &c.Logo); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return c, nil
}

func ListCompanyOutlets(companyID string) ([]CompanyOutlet, error) {
    rows, err := db.DB.Query(`SELECT id, company_id, name, address FROM company_outlets WHERE company_id = ?`, companyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var result []CompanyOutlet
    for rows.Next() {
        var o CompanyOutlet
        if err := rows.Scan(&o.ID, &o.CompanyID, &o.Name, &o.Address); err != nil {
            return nil, err
        }
        result = append(result, o)
    }
    return result, rows.Err()
}


