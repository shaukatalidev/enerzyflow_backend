package companies

import (
	"database/sql"
	"enerzyflow_backend/internal/db"
)

func UpsertCompanyTx(tx *sql.Tx, c *Company) error {
	res, err := tx.Exec(`UPDATE companies SET name = $1, address = $2, logo = $3, updated_at = CURRENT_TIMESTAMP WHERE user_id = $4`, c.Name, c.Address, c.Logo, c.UserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		_, err = tx.Exec(`INSERT INTO companies (company_id, user_id, name, address, logo) VALUES ($1, $2, $3, $4, $5)`, c.CompanyID, c.UserID, c.Name, c.Address, c.Logo)
		return err
	}
	return nil
}

func GetLabelByIDAndCompanyID(labelID, companyID string) (*Label, error) {
	row := db.DB.QueryRow(`SELECT label_id, company_id, name, label_url FROM labels WHERE label_id = $1 AND company_id = $2`, labelID, companyID)
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
	if _, err := tx.Exec(`DELETE FROM company_outlets WHERE company_id = $1`, companyID); err != nil {
		return err
	}
	for _, o := range outlets {
		if _, err := tx.Exec(`INSERT INTO company_outlets (id, company_id, name, address) VALUES ($1, $2, $3, $4)`, o.ID, o.CompanyID, o.Name, o.Address); err != nil {
			return err
		}
	}
	return nil
}

func GetCompanyByUserID(userID string) (*Company, error) {
	row := db.DB.QueryRow(`SELECT company_id, user_id, name, address, logo FROM companies WHERE user_id = $1`, userID)
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
	rows, err := db.DB.Query(`SELECT id, company_id, name, address FROM company_outlets WHERE company_id = $1`, companyID)
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
	rows, err := db.DB.Query(`SELECT label_id, company_id, name, label_url FROM labels WHERE company_id = $1 ORDER BY created_at DESC`, companyID)
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

func ReplaceCompanyLabelsTx(tx *sql.Tx, companyID string, labelsToSave []Label) ([]BlockedLabel, error) {
	var blocked []BlockedLabel
    rows, err := tx.Query(`SELECT label_id FROM labels WHERE company_id = $1`, companyID)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    existing := make(map[string]bool)
    for rows.Next() {
        var id string
        if err := rows.Scan(&id); err != nil {
            return nil, err
        }
        existing[id] = true
    }

	if len(labelsToSave) == 0 {
    for oldID := range existing {
        var count int
        if err := tx.QueryRow(`SELECT COUNT(*) FROM orders WHERE label_id = $1`, oldID).Scan(&count); err != nil {
            return nil, err
        }
        if count == 0 {
            if _, err := tx.Exec(`DELETE FROM labels WHERE label_id = $1`, oldID); err != nil {
                return nil, err
            }
        } else {
            var name string
            if err := tx.QueryRow(`SELECT name FROM labels WHERE label_id = $1`, oldID).Scan(&name); err != nil {
                return nil, err
            }
            blocked = append(blocked, BlockedLabel{
                LabelID: oldID,
                Name:    name,
            })
        }
    }
    return blocked, nil
}

    incoming := make(map[string]Label)
    for _, l := range labelsToSave {
        incoming[l.LabelID] = l
    }

    
    for oldID := range existing {
        if _, found := incoming[oldID]; !found {
            var count int
            if err := tx.QueryRow(`SELECT COUNT(*) FROM orders WHERE label_id = $1`, oldID).Scan(&count); err != nil {
                return nil, err
            }
            if count == 0 {
                if _, err := tx.Exec(`DELETE FROM labels WHERE label_id = $1`, oldID); err != nil {
                    return nil, err
                }
            } else {
                var name string
                if err := tx.QueryRow(`SELECT name FROM labels WHERE label_id = $1`, oldID).Scan(&name); err != nil {
                    return nil, err
                }
                blocked = append(blocked, BlockedLabel{
                    LabelID: oldID,
                    Name:    name,
                })
            }
        }
    }

    for _, l := range labelsToSave {
        if existing[l.LabelID] {
            if _, err := tx.Exec(
                `UPDATE labels SET name = $1, label_url = $2 WHERE label_id = $3 AND company_id = $4`,
                l.Name, l.URL, l.LabelID, companyID,
            ); err != nil {
                return blocked, err
            }
        } else {
			
            if _, err := tx.Exec(
                `INSERT INTO labels (label_id, company_id, name, label_url) VALUES ($1, $2, $3, $4)`,
                l.LabelID, companyID, l.Name, l.URL,
            ); err != nil {
                return blocked, err
            }
        }
    }

    return blocked, nil
}