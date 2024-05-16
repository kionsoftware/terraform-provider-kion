package kionclient

// AccountListResponse for GET /api/v3/account
type AccountListResponse struct {
	Data []struct {
		ID                        uint   `json:"id"`
		AccountNumber             string `json:"account_number"`
		Name                      string `json:"account_name"`
		Email                     string `json:"account_email"`
		LinkedRole                string `json:"linked_role"`
		ProjectID                 uint   `json:"project_id"`
		PayerID                   uint   `json:"payer_id"`
		AccountTypeID             uint   `json:"account_type_id"`
		StartDatecode             string `json:"start_datecode"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		UseOrgAccountInfo         bool   `json:"use_org_account_info"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		CARExternalID             string `json:"car_external_id"`
		ServiceExternalID         string `json:"service_external_id"`
		CreatedAt                 string `json:"created_at"`
	} `json:"data"`
	Status int `json:"status"`
}

// AccountResponse for: GET /api/v3/account/{id}
type AccountResponse struct {
	Data struct {
		ID                        uint   `json:"id"`
		AccountNumber             string `json:"account_number"`
		Name                      string `json:"account_name"`
		Email                     string `json:"account_email"`
		LinkedRole                string `json:"linked_role"`
		ProjectID                 uint   `json:"project_id"`
		PayerID                   uint   `json:"payer_id"`
		AccountTypeID             uint   `json:"account_type_id"`
		StartDatecode             string `json:"start_datecode"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		UseOrgAccountInfo         bool   `json:"use_org_account_info"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		CARExternalID             string `json:"car_external_id"`
		ServiceExternalID         string `json:"service_external_id"`
		CreatedAt                 string `json:"created_at"`
	}
	Status int `json:"status"`
}

func (r AccountResponse) ToMap(resource string) map[string]interface{} {
	accountNumberAttr := accountNumberAttr(resource)
	data := map[string]interface{}{
		accountNumberAttr:      r.Data.AccountNumber,
		"name":                 r.Data.Name,
		"project_id":           r.Data.ProjectID,
		"payer_id":             r.Data.PayerID,
		"account_type_id":      r.Data.AccountTypeID,
		"start_datecode":       r.Data.StartDatecode,
		"skip_access_checking": r.Data.SkipAccessChecking,
		"created_at":           r.Data.CreatedAt,
	}
	if resource == "kion_aws_account" {
		data["email"] = r.Data.Email
		data["linked_role"] = r.Data.LinkedRole
		data["linked_account_number"] = r.Data.LinkedAccountNumber
		data["include_linked_account_spend"] = r.Data.IncludeLinkedAccountSpend
		data["car_external_id"] = r.Data.CARExternalID
		data["service_external_id"] = r.Data.ServiceExternalID
		data["use_org_account_info"] = r.Data.UseOrgAccountInfo
	}
	return data
}

// AccountCacheListResponse for GET /api/v3/account-cache
type AccountCacheListResponse struct {
	Data []struct {
		ID                        uint   `json:"id"`
		AccountNumber             string `json:"account_number"`
		Name                      string `json:"account_name"`
		Email                     string `json:"account_email"`
		LinkedRole                string `json:"linked_role"`
		PayerID                   uint   `json:"payer_id"`
		AccountTypeID             uint   `json:"account_type_id"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		UseOrgAccountInfo         bool   `json:"use_org_account_info"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		CARExternalID             string `json:"car_external_id"`
		ServiceExternalID         string `json:"service_external_id"`
		CreatedAt                 string `json:"created_at"`
	} `json:"data"`
	Status int `json:"status"`
}

// AccountResponse for: GET /api/v3/account/cache/{id}
type AccountCacheResponse struct {
	Data struct {
		ID                        uint   `json:"id"`
		AccountNumber             string `json:"account_number"`
		Name                      string `json:"account_name"`
		Email                     string `json:"account_email"`
		LinkedRole                string `json:"linked_role"`
		PayerID                   uint   `json:"payer_id"`
		AccountTypeID             uint   `json:"account_type_id"`
		SkipAccessChecking        bool   `json:"skip_access_checking"`
		LinkedAccountNumber       string `json:"linked_account_number"`
		IncludeLinkedAccountSpend bool   `json:"include_linked_account_spend"`
		CARExternalID             string `json:"car_external_id"`
		ServiceExternalID         string `json:"service_external_id"`
		CreatedAt                 string `json:"created_at"`
	}
	Status int `json:"status"`
}

func (r AccountCacheResponse) ToMap(resource string) map[string]interface{} {
	accountNumberAttr := accountNumberAttr(resource)
	data := map[string]interface{}{
		accountNumberAttr:      r.Data.AccountNumber,
		"name":                 r.Data.Name,
		"payer_id":             r.Data.PayerID,
		"account_type_id":      r.Data.AccountTypeID,
		"skip_access_checking": r.Data.SkipAccessChecking,
		"created_at":           r.Data.CreatedAt,
		"project_id":           0,
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
	case "kion_aws_account":
		fallthrough
	default:
		return "account_number"
	}
}

// AccountCacheNewAWSCreate for: POST /api/v3/account-cache/create?account-type=aws
type AccountCacheNewAWSCreate struct {
	Name                      string                   `json:"account_name"`
	PayerID                   int                      `json:"payer_id"`
	AccountEmail              string                   `json:"account_email,omitempty"`
	CommercialAccountName     string                   `json:"commercial_account_name,omitempty"`
	GovAccountName            string                   `json:"gov_account_name,omitempty"`
	CreateGovcloud            *bool                    `json:"create_govcloud,omitempty"`
	IncludeLinkedAccountSpend *bool                    `json:"include_linked_account_spend,omitempty"`
	LinkedRole                string                   `json:"linked_role,omitempty"`
	OrganizationalUnit        *PayerOrganizationalUnit `json:"organizational_unit,omitempty"`
}

// AccountNewAWSImport for: POST /api/v3/account?account-type=aws
type AccountNewAWSImport struct {
	AccountEmail              string `json:"account_email,omitempty"`
	Name                      string `json:"account_name"`
	AccountNumber             string `json:"account_number"`
	AccountTypeID             *int   `json:"account_type_id,omitempty"`
	IncludeLinkedAccountSpend *bool  `json:"include_linked_account_spend,omitempty"`
	LinkedAccountNumber       string `json:"linked_aws_account_number,omitempty"`
	LinkedRole                string `json:"linked_role,omitempty"`
	PayerID                   int    `json:"payer_id"`
	ProjectID                 int    `json:"project_id"`
	SkipAccessChecking        *bool  `json:"skip_access_checking,omitempty"`
	StartDatecode             string `json:"start_datecode"`
	UseOrgAccountInfo         *bool  `json:"use_org_account_info,omitempty"`
}

// AccountCacheNewAWSImport for: POST /api/v3/account-cache?account-type=aws
type AccountCacheNewAWSImport struct {
	AccountEmail              string `json:"account_email,omitempty"`
	Name                      string `json:"account_name"`
	AccountNumber             string `json:"account_number"`
	AccountTypeID             *int   `json:"account_type_id,omitempty"`
	IncludeLinkedAccountSpend *bool  `json:"include_linked_account_spend,omitempty"`
	LinkedAccountNumber       string `json:"linked_aws_account_number,omitempty"`
	LinkedRole                string `json:"linked_role,omitempty"`
	PayerID                   int    `json:"payer_id"`
	SkipAccessChecking        *bool  `json:"skip_access_checking,omitempty"`
}

// PayerOrganizationalUnit represents an organizational unit in AWS payer's organization.
type PayerOrganizationalUnit struct {
	Name      string `json:"name"`
	OrgUnitId string `json:"org_unit_id"`
}

// AccountCacheNewGCPCreate for: POST /api/v3/account-cache/create?account-type=google-cloud
type AccountCacheNewGCPCreate struct {
	DisplayName           string `json:"display_name"`
	PayerID               int    `json:"payer_id"`
	GoogleCloudParentName string `json:"google_cloud_parent_name,omitempty"`
	GoogleCloudProjectID  string `json:"google_cloud_project_id,omitempty"`
}

// AccountNewGCPImport for: POST /api/v3/account?account-type=google-cloud
type AccountNewGCPImport struct {
	Name                 string `json:"account_name"`
	PayerID              int    `json:"payer_id"`
	AccountTypeID        *int   `json:"account_type_id,omitempty"`
	GoogleCloudProjectID string `json:"google_cloud_project_id"`
	SkipAccessChecking   *bool  `json:"skip_access_checking,omitempty"`
	ProjectID            int    `json:"project_id"`
	StartDatecode        string `json:"start_datecode"`
}

// AccountCacheNewGCPImport for: POST /api/v3/account-cache?account-type=google-cloud
type AccountCacheNewGCPImport struct {
	Name                 string `json:"account_name"`
	PayerID              int    `json:"payer_id"`
	AccountTypeID        *int   `json:"account_type_id,omitempty"`
	GoogleCloudProjectID string `json:"google_cloud_project_id"`
	SkipAccessChecking   *bool  `json:"skip_access_checking,omitempty"`
}

// AccountCacheNewAzureCreate for: POST /api/v3/account-cache/create?account-type=azure
type AccountCacheNewAzureCreate struct {
	Name                       string                      `json:"account_name"`
	SubscriptionName           string                      `json:"name"`
	ParentManagementGroupID    string                      `json:"parent_management_group_id,omitempty"`
	PayerID                    int                         `json:"payer_id"`
	SubscriptionEABillingInfo  *SubscriptionEABillingInfo  `json:"ea,omitempty"`
	SubscriptionMCABillingInfo *SubscriptionMCABillingInfo `json:"mca,omitempty"`
	SubscriptionCSPBillingInfo *SubscriptionCSPBillingInfo `json:"csp,omitempty"`
}

// AccountNewAzureImport for: POST /api/v3/account?account-type=azure
type AccountNewAzureImport struct {
	SubscriptionUUID   string `json:"subscription_uuid"`
	Name               string `json:"account_name"`
	ProjectID          int    `json:"project_id"`
	PayerID            int    `json:"payer_id"`
	StartDatecode      string `json:"start_datecode"`
	SkipAccessChecking *bool  `json:"skip_access_checking,omitempty"`
	AccountTypeID      *int   `json:"account_type_id,omitempty"`
}

// AccountCacheNewAzureImport for: POST /api/v3/account-cache?account-type=azure
type AccountCacheNewAzureImport struct {
	SubscriptionUUID   string `json:"subscription_uuid"`
	ResourceGroupName  string `json:"resource_group_name,omitempty"`
	Name               string `json:"account_name"`
	Email              string `json:"account_email,omitempty"`
	PayerID            int    `json:"payer_id"`
	SkipAccessChecking *bool  `json:"skip_access_checking,omitempty"`
	AccountTypeID      *int   `json:"account_type_id,omitempty"`
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
	AccountEmail              string `json:"account_email,omitempty"`
	Name                      string `json:"account_name,omitempty"`
	LinkedRole                string `json:"linked_role,omitempty"`
	StartDatecode             string `json:"start_datecode,omitempty"`
	IncludeLinkedAccountSpend *bool  `json:"include_linked_account_spend,omitempty"`
	SkipAccessChecking        *bool  `json:"skip_access_checking,omitempty"`
	UseOrgAccountInfo         *bool  `json:"use_org_account_info,omitempty"`
	ResetNotificationTime     *bool  `json:"reset_notification_tyime,omitempty"`
}

// AccountCacheUpdatable for: PATCH /api/v3/account-account/{id}
type AccountCacheUpdatable struct {
	AccountEmail              string `json:"account_email,omitempty"`
	Name                      string `json:"account_name,omitempty"`
	IncludeLinkedAccountSpend *bool  `json:"include_linked_account_spend,omitempty"`
	LinkedRole                string `json:"linked_role,omitempty"`
	SkipAccessChecking        *bool  `json:"skip_access_checking,omitempty"`
}

// AccountMove for: POST /api/v3/account/{id}/move
type AccountMove struct {
	ProjectID        int    `json:"project_id"`
	FinancialSetting string `json:"financials"`
	MoveDate         int    `json:"move_datecode"`
}

// Account Types
type CSPAccountType int

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
)
