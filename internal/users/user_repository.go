package users

import (
    "database/sql"
    "enerzyflow_backend/internal/db"
    "errors"
)

func GetUserByEmail(email string) (*User, error) {
    row := db.DB.QueryRow("SELECT user_id, email, name, phone, designation, role, profile_url FROM users WHERE email = ?", email)
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
    row := db.DB.QueryRow("SELECT user_id, email, name, phone, designation, role, profile_url FROM users WHERE user_id = ?", userID)
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
    _, err := db.DB.Exec("UPDATE users SET name = ?, phone = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", name, phone, id)
    return err
}

func InsertUser(u *User) error {
    if u.Email == "" || u.Role == "" || u.UserID == "" {
        return errors.New("user_id, email and role are required")
    }
    _, err := db.DB.Exec(`INSERT INTO users (user_id, email, role) VALUES (?, ?, ?)`, u.UserID, u.Email, u.Role)
    return err
}

func UpdateUserProfileTx(tx *sql.Tx, u *User) error {
    if u == nil || u.UserID == "" {
        return errors.New("user is nil or user_id missing")
    }
    _, err := tx.Exec(`UPDATE users SET name = ?, phone = ?, designation = ?, profile_url = ?, updated_at = CURRENT_TIMESTAMP WHERE user_id = ?`, u.Name, u.Phone, u.Designation, u.ProfileURL, u.UserID)
    return err
}

