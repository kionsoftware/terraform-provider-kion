package kionclient

// BillingSourceGovcloud defines an AWS GovCloud billing source.
type BillingSourceGovcloud struct {
	AccountCreationEnabled bool   `json:"account_creation_enabled,omitempty"`
	AWSAccountNumber       string `json:"aws_account_number,omitempty"`
	CARExternalID          string `json:"car_external_id,omitempty"`
	ID                     int    `json:"id,omitempty"`
	Name                   string `json:"name,omitempty"`
	PayerID                int    `json:"payer_id,omitempty"`
	ServiceExternalID      string `json:"service_external_id,omitempty"`
}

// BillingSourceGovcloudCreate defines the fields needed to create an AWS GovCloud billing source.
type BillingSourceGovcloudCreate struct {
	AccountCreationEnabled bool   `json:"account_creation_enabled"`
	AWSAccountNumber       string `json:"aws_account_number"`
	Name                   string `json:"name"`
}

// BillingSourceGovcloudUpdate defines the fields that can be updated on an AWS GovCloud billing source.
type BillingSourceGovcloudUpdate struct {
	AccountCreationEnabled *bool  `json:"account_creation_enabled,omitempty"`
	Name                   string `json:"name,omitempty"`
}

// BillingSourceGovcloudResponse is the response structure for AWS GovCloud billing source.
type BillingSourceGovcloudResponse struct {
	Data   BillingSourceGovcloud `json:"data"`
	Status int                   `json:"status"`
}