package kionclient

// BillingSourceListResponse for: GET /api/v4/billing-source
type BillingSourceListResponse struct {
	Data struct {
		Items []BillingSource `json:"items"`
		Total int             `json:"total"`
	} `json:"data"`
	Status int `json:"status"`
}

// BillingSource defines the contents of a billing source
type BillingSource struct {
	ID                    uint                        `json:"id"`
	AccountCreation       bool                        `json:"account_creation"`
	AWSPayer              *interface{}                `json:"aws_payer,omitempty"`
	AzurePayer            *interface{}                `json:"azure_payer,omitempty"`
	GCPPayer              *GCPBillingAccountAugmented `json:"gcp_payer,omitempty"`
	OCIPayer              *OCIBillingSource           `json:"oci_payer,omitempty"`
	UseFocusReports       bool                        `json:"use_focus_reports"`
	UseProprietaryReports bool                        `json:"use_proprietary_reports"`
}
