package companies

type Company struct {
    CompanyID string
    UserID    string
    Name      string
    Address   string
    Logo      string
}

type CompanyOutlet struct {
    ID        string
    CompanyID string
    Name      string
    Address   string
}


type Label struct {
    LabelID   string `json:"label_id"`
    CompanyID string `json:"company_id"`
    Name      string `json:"name"`
    URL       string `json:"label_url" db:"label_url"`
}

type LabelResponse struct {
    LabelID   string `json:"label_id"`
    Name      string `json:"name"`
    URL       string `json:"label_url"`
}

type BlockedLabel struct {
    LabelID string `json:"label_id"`
    Name    string `json:"name"`
}