package companies

import (
	"database/sql"
	"errors"
)

func SaveCompanyOutletsService(tx *sql.Tx, companyID string, outlets []CompanyOutlet) error {
	if companyID == "" {
		return errors.New("company_id cannot be empty")
	}

	for _, o := range outlets {
		if o.ID == "" {
			return errors.New("outlet ID cannot be empty")
		}
		if o.Name == "" {
			return errors.New("outlet name cannot be empty")
		}
		if o.Address == "" {
			return errors.New("outlet address cannot be empty")
		}
	}

	return ReplaceCompanyOutletsTx(tx, companyID, outlets)
}

func SaveCompanyLabelsService(tx *sql.Tx, companyID string, labels []Label) ([]BlockedLabel, error) {
	if companyID == "" {
		return nil, errors.New("company_id cannot be empty")
	}

	for _, l := range labels {
		if l.LabelID == "" {
			return nil, errors.New("label_id cannot be empty")
		}
		if l.Name == "" {
			return nil, errors.New("label name cannot be empty")
		}
		if l.URL == "" {
			return nil, errors.New("label URL cannot be empty")
		}
	}

	return ReplaceCompanyLabelsTx(tx, companyID, labels)
}