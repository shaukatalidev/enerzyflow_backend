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
    LabelID   string
    CompanyID string
    Name      string
    URL       string
}


