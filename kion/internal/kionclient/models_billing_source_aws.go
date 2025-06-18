package kionclient

// AWSBillingSourceCreate for creating AWS billing sources
type AWSBillingSourceCreate struct {
	Name                             string  `json:"name"`
	AWSAccountNumber                 string  `json:"aws_account_number"`
	AccountTypeID                    int     `json:"account_type_id"`
	AccountCreation                  bool    `json:"account_creation"`
	BillingBucketAccountNumber       string  `json:"billing_bucket_account_number,omitempty"`
	BillingRegion                    string  `json:"billing_region,omitempty"`
	BillingReportType                string  `json:"billing_report_type,omitempty"`
	BillingStartDate                 string  `json:"billing_start_date"`
	BucketAccessRole                 string  `json:"bucket_access_role,omitempty"`
	CURBucket                        string  `json:"cur_bucket,omitempty"`
	CURBucketRegion                  string  `json:"cur_bucket_region,omitempty"`
	CURName                          string  `json:"cur_name,omitempty"`
	CURPrefix                        string  `json:"cur_prefix,omitempty"`
	FocusBillingBucketAccountNumber  string  `json:"focus_billing_bucket_account_number,omitempty"`
	FocusBillingReportBucket         string  `json:"focus_billing_report_bucket,omitempty"`
	FocusBillingReportBucketRegion   string  `json:"focus_billing_report_bucket_region,omitempty"`
	FocusBillingReportName           string  `json:"focus_billing_report_name,omitempty"`
	FocusBillingReportPrefix         string  `json:"focus_billing_report_prefix,omitempty"`
	FocusBucketAccessRole            string  `json:"focus_bucket_access_role,omitempty"`
	KeyID                            string  `json:"key_id,omitempty"`
	KeySecret                        string  `json:"key_secret,omitempty"`
	LinkedRole                       string  `json:"linked_role,omitempty"`
	MRBucket                         string  `json:"mr_bucket,omitempty"`
	SkipValidation                   bool    `json:"skip_validation,omitempty"`
}

// AWSBillingSourceUpdate for updating AWS billing sources
type AWSBillingSourceUpdate struct {
	Name                             string  `json:"name,omitempty"`
	AccountCreation                  *bool   `json:"account_creation,omitempty"`
	BillingBucketAccountNumber       string  `json:"billing_bucket_account_number,omitempty"`
	BillingRegion                    string  `json:"billing_region,omitempty"`
	BillingReportType                string  `json:"billing_report_type,omitempty"`
	BillingStartDate                 string  `json:"billing_start_date,omitempty"`
	BucketAccessRole                 string  `json:"bucket_access_role,omitempty"`
	CURBucket                        string  `json:"cur_bucket,omitempty"`
	CURBucketRegion                  string  `json:"cur_bucket_region,omitempty"`
	CURName                          string  `json:"cur_name,omitempty"`
	CURPrefix                        string  `json:"cur_prefix,omitempty"`
	FocusBillingBucketAccountNumber  string  `json:"focus_billing_bucket_account_number,omitempty"`
	FocusBillingReportBucket         string  `json:"focus_billing_report_bucket,omitempty"`
	FocusBillingReportBucketRegion   string  `json:"focus_billing_report_bucket_region,omitempty"`
	FocusBillingReportName           string  `json:"focus_billing_report_name,omitempty"`
	FocusBillingReportPrefix         string  `json:"focus_billing_report_prefix,omitempty"`
	FocusBucketAccessRole            string  `json:"focus_bucket_access_role,omitempty"`
	KeyID                            string  `json:"key_id,omitempty"`
	KeySecret                        string  `json:"key_secret,omitempty"`
	LinkedRole                       string  `json:"linked_role,omitempty"`
	MRBucket                         string  `json:"mr_bucket,omitempty"`
}

// AWSPayer represents AWS billing source details
type AWSPayer struct {
	ID                              uint   `json:"id"`
	Name                            string `json:"name"`
	AccountNumber                   string `json:"account_number"`
	BillingBucketAccountNumber      string `json:"billing_bucket_account_number"`
	BillingRegion                   string `json:"billing_region"`
	BillingReportBucket             string `json:"billing_report_bucket"`
	BillingReportBucketRegion       string `json:"billing_report_bucket_region"`
	BillingReportName               string `json:"billing_report_name"`
	BillingReportPrefix             string `json:"billing_report_prefix"`
	BillingReportType               string `json:"billing_report_type"`
	BillingStartDate                string `json:"billing_start_date"`
	BucketAccessRole                string `json:"bucket_access_role"`
	DetailedBillingBucket           string `json:"detailed_billing_bucket"`
	FOCUSBillingBucketAccountNumber string `json:"focus_billing_bucket_account_number"`
	FOCUSBillingReportBucket        string `json:"focus_billing_report_bucket"`
	FOCUSBillingReportBucketRegion  string `json:"focus_billing_report_bucket_region"`
	FOCUSBillingReportName          string `json:"focus_billing_report_name"`
	FOCUSBillingReportPrefix        string `json:"focus_billing_report_prefix"`
	FOCUSBucketAccessRole           string `json:"focus_bucket_access_role"`
	OrgID                           string `json:"org_id"`
}

// BillingSourceResponse for billing source GET responses
type BillingSourceResponse struct {
	Data   BillingSourceData `json:"data"`
	Status int               `json:"status"`
}

// BillingSourceData contains the billing source details
type BillingSourceData struct {
	ID                    uint      `json:"id"`
	AccountCreation       bool      `json:"account_creation"`
	AWSPayer              *AWSPayer `json:"aws_payer,omitempty"`
	UseFocusReports       bool      `json:"use_focus_reports"`
	UseProprietaryReports bool      `json:"use_proprietary_reports"`
}