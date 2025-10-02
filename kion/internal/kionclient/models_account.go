package kionclient

// AccountListResponse for GET /api/v3/account
type AccountListResponse struct {
	Data []struct {
		AccountAlias              string `json:"account_alias"`
		AccountNumber             string `json:"account_number"`
		AccountTypeID             uint   `json:"account_type_id"`
		CARExternalID             string `json:"car_external_id"`
		CreatedAt                 string `json:"created_at"`
		Email                     string `json:"account_email"`
		ID                        uint   `json:"id"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		LinkedRole                string `json:"linked_role"`
		Name                      string `json:"account_name"`
		PayerID                   uint   `json:"payer_id"`
		ProjectID                 uint   `json:"project_id"`
		ServiceExternalID         string `json:"service_external_id"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		StartDatecode             string `json:"start_datecode"`
		UseOrgAccountInfo         bool   `json:"use_org_account_info"`
	} `json:"data"`
	Status int `json:"status"`
}

// AccountResponse for: GET /api/v3/account/{id}
type AccountResponse struct {
	Data struct {
		AccountAlias              string `json:"account_alias"`
		AccountNumber             string `json:"account_number"`
		AccountTypeID             uint   `json:"account_type_id"`
		CARExternalID             string `json:"car_external_id"`
		CreatedAt                 string `json:"created_at"`
		Email                     string `json:"account_email"`
		ID                        uint   `json:"id"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		LinkedRole                string `json:"linked_role"`
		Name                      string `json:"account_name"`
		PayerID                   uint   `json:"payer_id"`
		ProjectID                 uint   `json:"project_id"`
		ServiceExternalID         string `json:"service_external_id"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		StartDatecode             string `json:"start_datecode"`
		UseOrgAccountInfo         bool   `json:"use_org_account_info"`
	}
	Status int `json:"status"`
}

func (r AccountResponse) ToMap(resource string) map[string]interface{} {
	accountNumberAttr := accountNumberAttr(resource)
	data := map[string]interface{}{
		"account_type_id":      r.Data.AccountTypeID,
		accountNumberAttr:      r.Data.AccountNumber,
		"account_alias":        r.Data.AccountAlias,
		"created_at":           r.Data.CreatedAt,
		"name":                 r.Data.Name,
		"payer_id":             r.Data.PayerID,
		"project_id":           r.Data.ProjectID,
		"skip_access_checking": r.Data.SkipAccessChecking,
		"start_datecode":       r.Data.StartDatecode,
	}
	if resource == "kion_aws_account" {
		data["car_external_id"] = r.Data.CARExternalID
		data["email"] = r.Data.Email
		data["include_linked_account_spend"] = r.Data.IncludeLinkedAccountSpend
		data["linked_account_number"] = r.Data.LinkedAccountNumber
		data["linked_role"] = r.Data.LinkedRole
		data["service_external_id"] = r.Data.ServiceExternalID
		data["use_org_account_info"] = r.Data.UseOrgAccountInfo
	}
	return data
}

// AccountCacheListResponse for GET /api/v3/account-cache
type AccountCacheListResponse struct {
	Data []struct {
		AccountAlias              *string `json:"account_alias"`
		AccountNumber             string  `json:"account_number"`
		AccountTypeID             uint    `json:"account_type_id"`
		CARExternalID             string  `json:"car_external_id"`
		CreatedAt                 string  `json:"created_at"`
		Email                     string  `json:"account_email"`
		ID                        uint    `json:"id"`
		IncludeLinkedAccountSpend bool    `json:"include_linked_account_spend"`
		LinkedAccountNumber       string  `json:"linked_account_number"`
		LinkedRole                string  `json:"linked_role"`
		Name                      string  `json:"account_name"`
		PayerID                   uint    `json:"payer_id"`
		ServiceExternalID         string  `json:"service_external_id"`
		SkipAccessChecking        bool    `json:"skip_access_checking"`
		UseOrgAccountInfo         bool    `json:"use_org_account_info"`
	} `json:"data"`
	Status int `json:"status"`
}

// AccountResponse for: GET /api/v3/account/cache/{id}
type AccountCacheResponse struct {
	Data struct {
		AccountAlias              *string `json:"account_alias"`
		AccountNumber             string  `json:"account_number"`
		AccountTypeID             uint    `json:"account_type_id"`
		CARExternalID             string  `json:"car_external_id"`
		CreatedAt                 string  `json:"created_at"`
		Email                     string  `json:"account_email"`
		ID                        uint    `json:"id"`
		IncludeLinkedAccountSpend bool    `json:"include_linked_account_spend"`
		LinkedAccountNumber       string  `json:"linked_account_number"`
		LinkedRole                string  `json:"linked_role"`
		Name                      string  `json:"account_name"`
		PayerID                   uint    `json:"payer_id"`
		ServiceExternalID         string  `json:"service_external_id"`
		SkipAccessChecking        bool    `json:"skip_access_checking"`
	}
	Status int `json:"status"`
}

func (r AccountCacheResponse) ToMap(resource string) map[string]interface{} {
	accountNumberAttr := accountNumberAttr(resource)
	data := map[string]interface{}{
		"account_alias":        r.Data.AccountAlias,
		accountNumberAttr:      r.Data.AccountNumber,
		"account_type_id":      r.Data.AccountTypeID,
		"created_at":           r.Data.CreatedAt,
		"name":                 r.Data.Name,
		"payer_id":             r.Data.PayerID,
		"project_id":           0,
		"skip_access_checking": r.Data.SkipAccessChecking,
	}
	if resource == "kion_aws_account" {
		data["email"] = r.Data.Email
		data["linked_role"] = r.Data.LinkedRole
		data["linked_account_number"] = r.Data.LinkedAccountNumber
		data["include_linked_account_spend"] = r.Data.IncludeLinkedAccountSpend
		data["car_external_id"] = r.Data.CARExternalID
		data["service_external_id"] = r.Data.ServiceExternalID
	}
	return data
}

func accountNumberAttr(resource string) string {
	switch resource {
	case "kion_gcp_account":
		return "google_cloud_project_id"
	case "kion_azure_account":
		return "subscription_uuid"
	case "kion_custom_account":
		return "account_number"
	case "kion_aws_account":
		fallthrough
	default:
		return "account_number"
	}
}

// AccountCacheNewAWSCreate for: POST /api/v3/account-cache/create?account-type=aws
type AccountCacheNewAWSCreate struct {
	AccountAlias              *string                  `json:"account_alias,omitempty"`
	AccountEmail              string                   `json:"account_email,omitempty"`
	CommercialAccountName     string                   `json:"commercial_account_name,omitempty"`
	CreateGovcloud            *bool                    `json:"create_govcloud,omitempty"`
	GovAccountName            string                   `json:"gov_account_name,omitempty"`
	IncludeLinkedAccountSpend *bool                    `json:"include_linked_account_spend,omitempty"`
	LinkedRole                string                   `json:"linked_role,omitempty"`
	Name                      string                   `json:"account_name"`
	OrganizationalUnit        *PayerOrganizationalUnit `json:"organizational_unit,omitempty"`
	PayerID                   int                      `json:"payer_id"`
}

// AccountNewAWSImport for: POST /api/v3/account?account-type=aws
type AccountNewAWSImport struct {
	AccountAlias              *string `json:"account_alias,omitempty"`
	AccountEmail              string  `json:"account_email,omitempty"`
	AccountNumber             string  `json:"account_number"`
	AccountTypeID             *int    `json:"account_type_id,omitempty"`
	IncludeLinkedAccountSpend *bool   `json:"include_linked_account_spend,omitempty"`
	LinkedAccountNumber       string  `json:"linked_aws_account_number,omitempty"`
	LinkedRole                string  `json:"linked_role,omitempty"`
	Name                      string  `json:"account_name"`
	PayerID                   int     `json:"payer_id"`
	ProjectID                 int     `json:"project_id"`
	SkipAccessChecking        *bool   `json:"skip_access_checking,omitempty"`
	StartDatecode             string  `json:"start_datecode"`
	UseOrgAccountInfo         *bool   `json:"use_org_account_info,omitempty"`
}

// AccountCacheNewAWSImport for: POST /api/v3/account-cache?account-type=aws
type AccountCacheNewAWSImport struct {
	AccountAlias              *string `json:"account_alias,omitempty"`
	AccountEmail              string  `json:"account_email,omitempty"`
	AccountNumber             string  `json:"account_number"`
	AccountTypeID             *int    `json:"account_type_id,omitempty"`
	IncludeLinkedAccountSpend *bool   `json:"include_linked_account_spend,omitempty"`
	LinkedAccountNumber       string  `json:"linked_aws_account_number,omitempty"`
	LinkedRole                string  `json:"linked_role,omitempty"`
	Name                      string  `json:"account_name"`
	PayerID                   int     `json:"payer_id"`
	SkipAccessChecking        *bool   `json:"skip_access_checking,omitempty"`
}

// PayerOrganizationalUnit represents an organizational unit in AWS payer's organization.
type PayerOrganizationalUnit struct {
	Name      string `json:"name"`
	OrgUnitId string `json:"org_unit_id"`
}

// AccountCacheNewGCPCreate for: POST /api/v3/account-cache/create?account-type=google-cloud
type AccountCacheNewGCPCreate struct {
	AccountAlias          *string `json:"account_alias,omitempty"`
	DisplayName           string  `json:"display_name"`
	GoogleCloudParentName string  `json:"google_cloud_parent_name,omitempty"`
	GoogleCloudProjectID  string  `json:"google_cloud_project_id,omitempty"`
	PayerID               int     `json:"payer_id"`
}

// AccountNewGCPImport for: POST /api/v3/account?account-type=google-cloud
type AccountNewGCPImport struct {
	AccountAlias         *string `json:"account_alias,omitempty"`
	AccountTypeID        *int    `json:"account_type_id,omitempty"`
	GoogleCloudProjectID string  `json:"google_cloud_project_id"`
	Name                 string  `json:"account_name"`
	PayerID              int     `json:"payer_id"`
	ProjectID            int     `json:"project_id"`
	SkipAccessChecking   *bool   `json:"skip_access_checking,omitempty"`
	StartDatecode        string  `json:"start_datecode"`
}

// AccountCacheNewGCPImport for: POST /api/v3/account-cache?account-type=google-cloud
type AccountCacheNewGCPImport struct {
	AccountAlias         *string `json:"account_alias,omitempty"`
	AccountTypeID        *int    `json:"account_type_id,omitempty"`
	GoogleCloudProjectID string  `json:"google_cloud_project_id"`
	Name                 string  `json:"account_name"`
	PayerID              int     `json:"payer_id"`
	SkipAccessChecking   *bool   `json:"skip_access_checking,omitempty"`
}

// AccountCacheNewAzureCreate for: POST /api/v3/account-cache/create?account-type=azure
type AccountCacheNewAzureCreate struct {
	AccountAlias               *string                     `json:"account_alias,omitempty"`
	Name                       string                      `json:"account_name"`
	ParentManagementGroupID    string                      `json:"parent_management_group_id,omitempty"`
	PayerID                    int                         `json:"payer_id"`
	SubscriptionCSPBillingInfo *SubscriptionCSPBillingInfo `json:"csp,omitempty"`
	SubscriptionEABillingInfo  *SubscriptionEABillingInfo  `json:"ea,omitempty"`
	SubscriptionMCABillingInfo *SubscriptionMCABillingInfo `json:"mca,omitempty"`
	SubscriptionName           string                      `json:"name"`
}

// AccountNewAzureImport for: POST /api/v3/account?account-type=azure
type AccountNewAzureImport struct {
	AccountAlias       *string `json:"account_alias,omitempty"`
	AccountTypeID      *int    `json:"account_type_id,omitempty"`
	Name               string  `json:"account_name"`
	PayerID            int     `json:"payer_id"`
	ProjectID          int     `json:"project_id"`
	SkipAccessChecking *bool   `json:"skip_access_checking,omitempty"`
	StartDatecode      string  `json:"start_datecode"`
	SubscriptionUUID   string  `json:"subscription_uuid"`
}

// AccountCacheNewAzureImport for: POST /api/v3/account-cache?account-type=azure
type AccountCacheNewAzureImport struct {
	AccountAlias       *string `json:"account_alias,omitempty"`
	AccountTypeID      *int    `json:"account_type_id,omitempty"`
	Email              string  `json:"account_email,omitempty"`
	Name               string  `json:"account_name"`
	PayerID            int     `json:"payer_id"`
	ResourceGroupName  string  `json:"resource_group_name,omitempty"`
	SkipAccessChecking *bool   `json:"skip_access_checking,omitempty"`
	SubscriptionUUID   string  `json:"subscription_uuid"`
}

type SubscriptionEABillingInfo struct {
	BillingAccountNumber string `json:"billing_account,omitempty"`
	EAAccountNumber      string `json:"account,omitempty"`
}

type SubscriptionMCABillingInfo struct {
	BillingAccountNumber string `json:"billing_account,omitempty"`
	BillingProfileNumber string `json:"billing_profile,omitempty"`
	InvoiceSectionNumber string `json:"billing_profile_invoice,omitempty"`
}

type SubscriptionCSPBillingInfo struct {
	BillingCycle string `json:"billing_cycle,omitempty"`
	OfferID      string `json:"offer_id"`
}

type AccountRevertResponse struct {
	Status   int    `json:"status"`
	RecordID int    `json:"record_id"`
	Message  string `json:"message"`
}

// AccountUpdatable for: PATCH /api/v3/account/{id}
type AccountUpdatable struct {
	AccountAlias              *string `json:"account_alias,omitempty"`
	AccountEmail              string  `json:"account_email,omitempty"`
	IncludeLinkedAccountSpend *bool   `json:"include_linked_account_spend,omitempty"`
	LinkedRole                string  `json:"linked_role,omitempty"`
	Name                      string  `json:"account_name,omitempty"`
	ResetNotificationTime     *bool   `json:"reset_notification_tyime,omitempty"`
	SkipAccessChecking        *bool   `json:"skip_access_checking,omitempty"`
	StartDatecode             string  `json:"start_datecode,omitempty"`
	UseOrgAccountInfo         *bool   `json:"use_org_account_info,omitempty"`
}

// AccountCacheUpdatable for: PATCH /api/v3/account-account/{id}
type AccountCacheUpdatable struct {
	AccountAlias              *string `json:"account_alias,omitempty"`
	AccountEmail              string  `json:"account_email,omitempty"`
	IncludeLinkedAccountSpend *bool   `json:"include_linked_account_spend,omitempty"`
	LinkedRole                string  `json:"linked_role,omitempty"`
	Name                      string  `json:"account_name,omitempty"`
	SkipAccessChecking        *bool   `json:"skip_access_checking,omitempty"`
}

// AccountMove for: POST /api/v3/account/{id}/move
type AccountMove struct {
	ProjectID        int    `json:"project_id"`
	FinancialSetting string `json:"financials"`
	MoveDate         int    `json:"move_datecode"`
}

// AccountNewCustomImport for: POST /api/v3/account?account-type=custom
type AccountNewCustomImport struct {
	AccountAlias  *string `json:"account_alias,omitempty"`
	AccountNumber string  `json:"account_number"`
	Name          string  `json:"account_name"`
	PayerID       int     `json:"payer_id"`
	ProjectID     int     `json:"project_id"`
	StartDatecode string  `json:"start_datecode"`
}

// AccountCacheNewCustomImport for: POST /api/v3/account-cache?account-type=custom
type AccountCacheNewCustomImport struct {
	AccountAlias  *string `json:"account_alias,omitempty"`
	AccountNumber string  `json:"account_number"`
	Name          string  `json:"account_name"`
	PayerID       int     `json:"payer_id"`
}

// Account Types
type CSPAccountType uint

const (
	// AWSStandard is a Standard AWS account
	AWSStandard CSPAccountType = 1

	// AWSGovCloud is an AWS GovCloud account
	AWSGovCloud CSPAccountType = 2

	// AzureCSPStandard is a Standard Azure CSP subscription
	AzureCSPStandard CSPAccountType = 3

	// AWSC2S is an AWS account in the c2s region
	AWSC2S CSPAccountType = 4

	// AWSSC2S is an AWS account in the sc2s region
	AWSSC2S CSPAccountType = 5

	// AzureEA is an Azure EA subscription
	AzureEA CSPAccountType = 6

	// AzureEAGov is a subscription in Azure Gov backed by an EA
	AzureEAGov CSPAccountType = 7

	// AzureCSPStandardRG is a Azure CSP resource group
	AzureCSPStandardRG CSPAccountType = 8

	// AzureEARG is an Azure EA resource group
	AzureEARG CSPAccountType = 9

	// AzureEAGovRG is a resource group in Azure Gov backed by an EA
	AzureEAGovRG CSPAccountType = 10

	// AzureCSPGov is a subscription billed via an Azure CSP in a gov region
	AzureCSPGov CSPAccountType = 11

	// AzureCSPGovRG is a resource group in a subscription billed via an Azure CSP in a gov region
	AzureCSPGovRG CSPAccountType = 12

	// AzureEASecret is a subscription billed via Azure EA in Azure secret
	AzureEASecret CSPAccountType = 13

	// AzureEASecretRG is a resource group in a subscription billed via Azure EA in Azure secret
	AzureEASecretRG CSPAccountType = 14

	// GoogleCloudStandard is a Google Cloud project
	GoogleCloudStandard CSPAccountType = 15

	// AzureMCA is a Azure MCA subscription
	AzureMCA CSPAccountType = 16

	// AzureMCARG is a Azure MCA resource group
	AzureMCARG CSPAccountType = 17

	// AzureMCAGov is a subscription billed via an Azure MCA in a gov region
	AzureMCAGov CSPAccountType = 18

	// AzureMCAGovRG is a resource group in a subscription billed via an Azure MCA in a gov region
	AzureMCAGovRG CSPAccountType = 19

	AzureCSPTopSecret   CSPAccountType = 20
	AzureCSPTopSecretRG CSPAccountType = 21
	AzureEATopSecret    CSPAccountType = 22
	AzureEATopSecretRG  CSPAccountType = 23
	AzureMCATopSecret   CSPAccountType = 24
	AzureMCATopSecretRG CSPAccountType = 25
	OCICommercial       CSPAccountType = 26
	OCIGov              CSPAccountType = 27
	OCIFederal          CSPAccountType = 28

	// CustomAccount is a custom account type, not associated with any cloud provider
	CustomAccount CSPAccountType = 29
)
