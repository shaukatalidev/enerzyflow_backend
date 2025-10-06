package users

import (
    "database/sql"
    "enerzyflow_backend/internal/db"
    "errors"
)

func GetUserByEmail(email string) (*User, error) {
    row := db.DB.QueryRow("SELECT user_id, email, name, phone, designation, role, profile_url FROM users WHERE email = $1", email)
    u := &User{}
    if err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.Phone, &u.Designation, &u.Role, &u.ProfileURL); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return u, nil
}

func GetUserByPhone(phone string) (*User, error) {
    row:= db.DB.QueryRow("SELECT user_id, email, name, phone, designation, role, profile_url FROM users WHERE phone = $1", phone)
    u := &User{}
    if err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.Phone, &u.Designation, &u.Role, &u.ProfileURL); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return u, nil
}

func GetUserByID(userID string) (*User, error) {
    row := db.DB.QueryRow("SELECT user_id, email, name, COALESCE(phone, ''), designation, role, profile_url FROM users WHERE user_id = $1", userID)
    u := &User{}
    if err := row.Scan(&u.UserID, &u.Email, &u.Name, &u.Phone, &u.Designation, &u.Role, &u.ProfileURL); err != nil {
        if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return u, nil
}

func UpdateUserProfile(id int, name string, phone string) error {
    _, err := db.DB.Exec("UPDATE users SET name = $1, phone = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3", name, phone, id)
    return err
}

func InsertUser(u *User) error {
    if u.Email == "" || u.Role == "" || u.UserID == "" {
        return errors.New("user_id, email and role are required")
    }
    _, err := db.DB.Exec(`INSERT INTO users (user_id, email, role) VALUES ($1, $2, $3)`, u.UserID, u.Email, u.Role)
    return err
}

func UpdateUserProfileTx(tx *sql.Tx, u *User) error {
    if u == nil || u.UserID == "" {
        return errors.New("user is nil or user_id missing")
    }
    _, err := tx.Exec(`UPDATE users SET name = $1, phone = $2, designation = $3, profile_url = $4, updated_at = CURRENT_TIMESTAMP WHERE user_id = $5`, u.Name, u.Phone, u.Designation, u.ProfileURL, u.UserID)
    return err
}

func GetAllUsers() ([]User, error) {
	rows, err := db.DB.Query(`SELECT user_id, email, name, phone, role, profile_url FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var usersList []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.UserID, &u.Email, &u.Name, &u.Phone, &u.Role, &u.ProfileURL); err != nil {
			return nil, err
		}
		usersList = append(usersList, u)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return usersList, nil
}