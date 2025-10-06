package users

import (
	"database/sql"
	"enerzyflow_backend/internal/companies"
	"enerzyflow_backend/internal/db"
	"errors"
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
	resp := &SaveProfileResponse{}
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

	if req.Profile.Phone != "" {
		existingUser, err := GetUserByPhone(req.Profile.Phone)
		if err != nil {
			return nil, err
		}
		if existingUser != nil && existingUser.UserID != authenticatedUserID {
			return nil, errors.New("phone number already in use by another user")
		}
	}

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
		CompanyID: func() string {
			if existingCompany != nil {
				return existingCompany.CompanyID
			}
			return uuid.New().String()
		}(),
		UserID:  u.UserID,
		Name:    req.Company.Name,
		Address: req.Company.Address,
		Logo:    req.Company.Logo,
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
	if err = companies.SaveCompanyOutletsService(tx, company.CompanyID, outlets); err != nil {
		return nil, err
	}

	labelsToSave := make([]companies.Label, 0, len(req.Labels))
	for _, l := range req.Labels {
		id := l.LabelID
		if id == "" {
			id = uuid.New().String()
		}

		label := companies.Label{
			LabelID:   id,
			CompanyID: company.CompanyID,
			Name:      l.Name,
			URL:       l.URL,
		}

		labelsToSave = append(labelsToSave, label)

		resp.Labels = append(resp.Labels, companies.LabelResponse{
			LabelID: id,
			Name:    l.Name,
			URL:     l.URL,
		})
	}

	blocked, err := companies.SaveCompanyLabelsService(tx, company.CompanyID, labelsToSave)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

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
	for _, o := range outlets {
		resp.Company.Outlets = append(resp.Company.Outlets, struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Address string `json:"address"`
		}{ID: o.ID, Name: o.Name, Address: o.Address})
	}
	resp.BlockedLabels = blocked

	return resp, nil
}

func GetProfileService(authenticatedUserID string) (*SaveProfileResponse, error) {
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
			resp.Labels = append(resp.Labels, companies.LabelResponse{
				LabelID: l.LabelID,
				Name:    l.Name,
				URL:     l.URL,
			})
		}
	}
	return resp, nil
}
