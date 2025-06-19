package kionclient

// OCIBillingSource represents an OCI billing source
type OCIBillingSource struct {
	ID                    uint   `json:"id,omitempty"`
	Name                  string `json:"name"`
	AccountTypeID         uint   `json:"account_type_id"`
	BillingStartDate      string `json:"billing_start_date"`
	TenancyOCID           string `json:"tenancy_ocid"`
	UserOCID              string `json:"user_ocid"`
	Fingerprint           string `json:"fingerprint"`
	PrivateKey            string `json:"private_key"`
	Region                string `json:"region"`
	IsParentTenancy       bool   `json:"is_parent_tenancy"`
	UseFOCUSReports       bool   `json:"use_focus_reports"`
	UseProprietaryReports bool   `json:"use_proprietary_reports"`
}

// OCIBillingSourceCreate represents the payload for creating an OCI billing source
type OCIBillingSourceCreate struct {
	Name                  string `json:"name"`
	AccountTypeID         uint   `json:"account_type_id"`
	BillingStartDate      string `json:"billing_start_date"`
	TenancyOCID           string `json:"tenancy_ocid,omitempty"`
	UserOCID              string `json:"user_ocid,omitempty"`
	Fingerprint           string `json:"fingerprint,omitempty"`
	PrivateKey            string `json:"private_key,omitempty"`
	Region                string `json:"region,omitempty"`
	IsParentTenancy       bool   `json:"is_parent_tenancy"`
	UseFOCUSReports       bool   `json:"use_focus_reports"`
	UseProprietaryReports bool   `json:"use_proprietary_reports"`
	SkipValidation        bool   `json:"skip_validation,omitempty"`
}

// OCIBillingSourceUpdate represents the payload for updating an OCI billing source
type OCIBillingSourceUpdate struct {
	ID                    uint   `json:"id,omitempty"`
	Name                  string `json:"name,omitempty"`
	AccountTypeID         uint   `json:"account_type_id,omitempty"`
	BillingStartDate      string `json:"billing_start_date,omitempty"`
	TenancyOCID           string `json:"tenancy_ocid,omitempty"`
	UserOCID              string `json:"user_ocid,omitempty"`
	Fingerprint           string `json:"fingerprint,omitempty"`
	PrivateKey            string `json:"private_key,omitempty"`
	Region                string `json:"region,omitempty"`
	IsParentTenancy       bool   `json:"is_parent_tenancy"`
	UseFOCUSReports       bool   `json:"use_focus_reports"`
	UseProprietaryReports bool   `json:"use_proprietary_reports"`
	SkipValidation        bool   `json:"skip_validation,omitempty"`
}
