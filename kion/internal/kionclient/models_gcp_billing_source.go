package kionclient

// GCPBillingSourceCreate for: POST /v3/billing-source/gcp
type GCPBillingSourceCreate struct {
	AccountTypeID              uint                       `json:"account_type_id"`
	GCPBillingAccountCreate    GCPBillingAccountWithStart `json:"gcp_billing_account_create"`
}

// GCPBillingAccountWithStart contains fields describing a billing account in GCP with start date
type GCPBillingAccountWithStart struct {
	ServiceAccountID   uint               `json:"service_account_id"`
	Name               string             `json:"name"`
	GCPID              string             `json:"gcp_id"`
	BigQueryExport     GCPBigQueryExport  `json:"big_query_export"`
	BillingStartDate   string             `json:"billing_start_date"`
	IsReseller         bool               `json:"is_reseller,omitempty"`
	UseFOCUS           bool               `json:"use_focus,omitempty"`
	UseProprietary     bool               `json:"use_proprietary,omitempty"`
}

// GCPBigQueryExport describes information about where billing data is exported in BigQuery
type GCPBigQueryExport struct {
	GCPProjectID    string `json:"gcp_project_id"`
	DatasetName     string `json:"dataset_name"`
	TableName       string `json:"table_name"`
	TableFormat     string `json:"table_format,omitempty"`
	FOCUSViewName   string `json:"focus_view_name,omitempty"`
}

// GCPBillingAccount contains fields describing a billing account in GCP
type GCPBillingAccount struct {
	ID                uint              `json:"id"`
	ServiceAccountID  uint              `json:"service_account_id"`
	Name              string            `json:"name"`
	GCPID             string            `json:"gcp_id"`
	BigQueryExport    GCPBigQueryExport `json:"big_query_export"`
	IsReseller        bool              `json:"is_reseller"`
}

// GCPBillingAccountAugmented contains information about retrieving billing data from GCP
type GCPBillingAccountAugmented struct {
	GCPBillingAccount   GCPBillingAccount                `json:"gcp_billing_account"`
	GCPServiceAccount   GoogleCloudServiceAccountWithKey `json:"gcp_service_account"`
}

// GoogleCloudServiceAccountWithKey represents a GCP service account with key information
type GoogleCloudServiceAccountWithKey struct {
	ID                      uint   `json:"id"`
	Name                    string `json:"name"`
	Description             string `json:"description"`
	Email                   string `json:"email"`
	EnableFederationSupport bool   `json:"enable_federation_support"`
	ProjectID               string `json:"project_id"`
	UniqueID                string `json:"unique_id"`
	OAuthClientID           string `json:"oauth_client_id,omitempty"`
	// Private key is sensitive and not returned in GET responses
}

// GoogleCloudServiceAccountCreate for creating a service account
type GoogleCloudServiceAccountCreate struct {
	Name                    string `json:"name"`
	Description             string `json:"description"`
	EnableFederationSupport bool   `json:"enable_federation_support"`
	JSON                    string `json:"json,omitempty"`
	OAuthClientID           string `json:"oauth_client_id,omitempty"`
	OAuthClientSecret       string `json:"oauth_client_secret,omitempty"`
}

// GCPBillingSource represents the full GCP billing source including all fields
type GCPBillingSource struct {
	ID                    uint                       `json:"id"`
	Name                  string                     `json:"name"`
	AccountTypeID         uint                       `json:"account_type_id"`
	ServiceAccountID      uint                       `json:"service_account_id"`
	GCPID                 string                     `json:"gcp_id"`
	BillingStartDate      string                     `json:"billing_start_date"`
	BigQueryExport        GCPBigQueryExport          `json:"big_query_export"`
	IsReseller            bool                       `json:"is_reseller"`
	UseFOCUSReports       bool                       `json:"use_focus_reports"`
	UseProprietaryReports bool                       `json:"use_proprietary_reports"`
}