package users

type User struct {
	UserID      string
	Email       string
	Name        string
	Phone       string
	Designation string
	Role        string
}

type SaveProfileRequest struct {
    Profile struct {
        UserID      string `json:"user_id"`
        Email       string `json:"email"`
        Name        string `json:"name"`
        Phone       string `json:"phone"`
        Designation string `json:"designation"`
    } `json:"profile"`
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
}

type SaveProfileResponse struct {
    User struct {
        UserID      string `json:"user_id"`
        Email       string `json:"email"`
        Name        string `json:"name"`
        Phone       string `json:"phone"`
        Designation string `json:"designation"`
        Role        string `json:"role"`
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
}


