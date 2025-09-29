package users

import (
    "database/sql"
    "errors"

    "enerzyflow_backend/internal/companies"
    "enerzyflow_backend/internal/db"
    "github.com/google/uuid"
)

func GetUserByEmailService(email string) (*User, bool, error) {
    var u *User
    u, err := GetUserByEmail(email)

    if err == sql.ErrNoRows {
        return nil, false, nil
    }
    if err != nil {
        return nil, false, err
    }

    return u, true, nil
}

func SaveProfileService(authenticatedUserID string, req SaveProfileRequest) (*SaveProfileResponse, error) {
    if authenticatedUserID == "" {
        return nil, errors.New("missing authenticated user id")
    }

    u, err := GetUserByID(authenticatedUserID)
    if err != nil {
        return nil, err
    }
    if u == nil {
        return nil, errors.New("user not found")
    }

    tx, err := db.DB.Begin()
    if err != nil {
        return nil, err
    }
    defer func() {
        if err != nil {
            _ = tx.Rollback()
        }
    }()

    u.Name = req.Profile.Name
    u.Phone = req.Profile.Phone
    u.Designation = req.Profile.Designation
    u.ProfileURL = req.Profile.ProfileURL
    if err = UpdateUserProfileTx(tx, u); err != nil {
        return nil, err
    }

    existingCompany, err := companies.GetCompanyByUserID(u.UserID)
    if err != nil {
        return nil, err
    }
    company := &companies.Company{
        CompanyID: func() string { if existingCompany != nil { return existingCompany.CompanyID }; return uuid.New().String() }(),
        UserID:    u.UserID,
        Name:      req.Company.Name,
        Address:   req.Company.Address,
        Logo:      req.Company.Logo,
    }
    if err = companies.UpsertCompanyTx(tx, company); err != nil {
        return nil, err
    }

    outlets := make([]companies.CompanyOutlet, 0, len(req.Company.Outlets))
    for _, o := range req.Company.Outlets {
        id := uuid.New().String()
        outlets = append(outlets, companies.CompanyOutlet{
            ID:        id,
            CompanyID: company.CompanyID,
            Name:      o.Name,
            Address:   o.Address,
        })
    }
    if err = companies.ReplaceCompanyOutletsTx(tx, company.CompanyID, outlets); err != nil {
        return nil, err
    }

    if len(req.Company.Labels) > 0 {
        labelsToSave := make([]companies.Label, 0, len(req.Company.Labels))
        for _, l := range req.Company.Labels {
            id := uuid.New().String()
            labelsToSave = append(labelsToSave, companies.Label{
                LabelID:   id,
                CompanyID: company.CompanyID,
                Name:      l.Name,
                URL:       l.URL,
            })
        }
        if err = companies.ReplaceCompanyLabelsTx(tx, company.CompanyID, labelsToSave); err != nil {
            return nil, err
        }
    }

    if err = tx.Commit(); err != nil {
        return nil, err
    }

    resp := &SaveProfileResponse{}
    resp.User.UserID = u.UserID
    resp.User.Email = u.Email
    resp.User.Name = u.Name
    resp.User.Phone = u.Phone
    resp.User.Designation = u.Designation
    resp.User.Role = u.Role
    resp.User.ProfileURL = u.ProfileURL
    resp.Company.CompanyID = company.CompanyID
    resp.Company.Name = company.Name
    resp.Company.Address = company.Address
    resp.Company.Logo = company.Logo
    // echo back saved outlets
    for _, o := range outlets {
        resp.Company.Outlets = append(resp.Company.Outlets, struct {
            ID      string `json:"id"`
            Name    string `json:"name"`
            Address string `json:"address"`
        }{ID: o.ID, Name: o.Name, Address: o.Address})
    }
    return resp, nil
}

func GetProfileService(authenticatedUserID string) (*SaveProfileResponse, error) {
    // if authenticatedUserID == "" {
    //     return nil, errors.New("missing authenticated user id")
    // }
    u, err := GetUserByID(authenticatedUserID)
    if err != nil {
        return nil, err
    }
    if u == nil {
        return nil, errors.New("user not found")
    }

    resp := &SaveProfileResponse{}
    resp.User.UserID = u.UserID
    resp.User.Email = u.Email
    resp.User.Name = u.Name
    resp.User.Phone = u.Phone
    resp.User.Designation = u.Designation
    resp.User.Role = u.Role
    resp.User.ProfileURL = u.ProfileURL

    company, err := companies.GetCompanyByUserID(u.UserID)
    if err != nil {
        return nil, err
    }
    if company != nil {
        resp.Company.CompanyID = company.CompanyID
        resp.Company.Name = company.Name
        resp.Company.Address = company.Address
        resp.Company.Logo = company.Logo
        outs, err := companies.ListCompanyOutlets(company.CompanyID)
        if err != nil {
            return nil, err
        }
        for _, o := range outs {
            resp.Company.Outlets = append(resp.Company.Outlets, struct {
                ID      string `json:"id"`
                Name    string `json:"name"`
                Address string `json:"address"`
            }{ID: o.ID, Name: o.Name, Address: o.Address})
        }
        
        companyLabels, err := companies.GetLabelsByCompanyID(company.CompanyID)
        if err != nil {
            return nil, err
        }
        for _, l := range companyLabels {
            resp.Labels = append(resp.Labels, struct {
                LabelID string `json:"label_id"`
                Name    string `json:"name"`
                URL     string `json:"label_url"`
            }{LabelID: l.LabelID, Name: l.Name, URL: l.URL})
        }
    }
    return resp, nil
}
