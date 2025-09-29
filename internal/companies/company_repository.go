package companies

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
)

func UpsertCompanyTx(tx *sql.Tx, c *Company) error {
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


func GetLabelByIDAndCompanyID(labelID, companyID string) (*Label, error) {
    row := db.DB.QueryRow(`SELECT label_id, company_id, name, label_url FROM labels WHERE label_id = ? AND company_id = ?`, labelID, companyID)
    var l Label
    if err := row.Scan(&l.LabelID, &l.CompanyID, &l.Name, &l.URL); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &l, nil
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


func GetLabelsByCompanyID(companyID string) ([]Label, error) {
    rows, err := db.DB.Query(`SELECT label_id, company_id, name, label_url FROM labels WHERE company_id = ? ORDER BY created_at DESC`, companyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var out []Label
    for rows.Next() {
        var l Label
        if err := rows.Scan(&l.LabelID, &l.CompanyID, &l.Name, &l.URL); err != nil {
            return nil, err
        }
        out = append(out, l)
    }
    return out, rows.Err()
}

func ReplaceCompanyLabelsTx(tx *sql.Tx, companyID string, labelsToSave []Label) error {
    if _, err := tx.Exec(`DELETE FROM labels WHERE company_id = ?`, companyID); err != nil {
        return err
    }
    if len(labelsToSave) == 0 {
        return nil
    }
    insertQuery := `INSERT INTO labels (label_id, company_id, name, label_url) VALUES (?, ?, ?, ?)`
    for _, l := range labelsToSave {
        if _, err := tx.Exec(insertQuery, l.LabelID, companyID, l.Name, l.URL); err != nil {
            return err
        }
    }
    return nil
}


