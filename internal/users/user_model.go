package users

import (
	"enerzyflow_backend/internal/companies"
)

type User struct {
	UserID      string
	Email       string
	Name        string
	Phone       *string
	Designation string
	Role        string
	ProfileURL  string
}

type SaveProfileRequest struct {
    Profile struct {
        UserID      string `json:"user_id"`
        Email       string `json:"email"`
        Name        string `json:"name"`
        Phone       *string `json:"phone"`
        Designation string `json:"designation"`
        ProfileURL  string `json:"profile_url"`
    } `json:"profile"`
    Company struct {
        CompanyID string `json:"company_id"`
        Name      string `json:"name"`
        Address   string `json:"address"`
        Logo      string `json:"logo_url"`
        Outlets   []struct {
            ID      string `json:"id"`
            Name    string `json:"name"`
            Address string `json:"address"`
        } `json:"outlets"`
    } `json:"company"`
    Labels []companies.Label `json:"labels"`
}

type SaveProfileResponse struct {
    User struct {
        UserID      string `json:"user_id"`
        Email       string `json:"email"`
        Name        string `json:"name"`
        Phone       *string `json:"phone"`
        Designation string `json:"designation"`
        Role        string `json:"role"`
        ProfileURL  string `json:"profile_url"`
    } `json:"user"`
    Company struct {
        CompanyID string `json:"company_id"`
        Name      string `json:"name"`
        Address   string `json:"address"`
        Logo      string `json:"logo"`
        Outlets   []struct {
            ID      string `json:"id"`
            Name    string `json:"name"`
            Address string `json:"address"`
        } `json:"outlets"`
    } `json:"company"`
    Labels []companies.LabelResponse `json:"labels"` 
    BlockedLabels []companies.BlockedLabel `json:"blocked_labels"`
}

type CreateUserRequest struct {
    Email string `json:"email" binding:"required,email"`
    Role  string `json:"role" binding:"required,oneof=printing plant"`
}