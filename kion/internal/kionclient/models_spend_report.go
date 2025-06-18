package kionclient

// SpendReportSearchRequest represents the request body for searching spend reports
type SpendReportSearchRequest struct {
	Query              string   `json:"query"`
	States             []string `json:"states"`
	StartDatecode      *int     `json:"start_datecode"`
	EndDatecode        *int     `json:"end_datecode"`
	UserIds            []int    `json:"user_ids"`
	UserGroupIds       []int    `json:"ugroup_ids"`
	IDMSIds            []int    `json:"idms_ids"`
	AppLabelIds        []int    `json:"app_label_ids"`
	CloudProviderTypeId *int    `json:"cloud_provider_type_id"`
	PartitionId        int      `json:"partition_id"`
	OUIds              []int    `json:"ou_ids"`
	ProjectIds         []int    `json:"project_ids"`
	AccountIds         []int    `json:"account_ids"`
	PayerIds           []int    `json:"payer_ids"`
	ItemType           string   `json:"item_type"`
}

// SpendReportListResponse represents the response from listing spend reports
type SpendReportListResponse struct {
	Status int `json:"status"`
	Data   struct {
		Pagination struct {
			Augmented  bool   `json:"augmented"`
			Page       int    `json:"page"`
			Count      int    `json:"count"`
			SortMethod string `json:"sort_method"`
			SortOrder  string `json:"sort_order"`
		} `json:"pagination"`
		Total    int                   `json:"total"`
		Items    []SpendReportListItem `json:"items"`
		Metadata struct {
			HiddenCount int `json:"hidden_count"`
		} `json:"metadata"`
	} `json:"data"`
}

// SpendReportListItem represents a single spend report in the list response
type SpendReportListItem struct {
	SavedReport SpendReport `json:"saved_report"`
	UserIds     []int       `json:"user_ids"`
	UserGroupIds []int      `json:"ugroup_ids"`
	Hidden      bool        `json:"hidden"`
}

// TimeField represents the Kion API time field structure
type TimeField struct {
	Time  string `json:"Time"`
	Valid bool   `json:"Valid"`
}

// SpendReport represents the core spend report structure
type SpendReport struct {
	ID                                    int                              `json:"id"`
	CreatedBy                             int                              `json:"created_by"`
	ReportName                            string                           `json:"report_name"`
	GlobalVisibility                      bool                             `json:"global_visibility"`
	DateRange                             string                           `json:"date_range"`
	Scope                                 string                           `json:"scope"`
	ScopeId                               int                              `json:"scope_id"`
	StartMonth                            int                              `json:"start_month,omitempty"`
	EndMonth                              int                              `json:"end_month,omitempty"`
	StartDate                             string                           `json:"start_date,omitempty"`
	EndDate                               string                           `json:"end_date,omitempty"`
	SpendType                             string                           `json:"spend_type"`
	Dimension                             string                           `json:"dimension"`
	TimeGranularityId                     int                              `json:"time_granularity_id"`
	OUIds                                 []int                            `json:"ou_ids"`
	OUExclusive                           bool                             `json:"ou_exclusive"`
	IncludeDescendants                    bool                             `json:"include_descendants"`
	ProjectIds                            []int                            `json:"project_ids"`
	ProjectExclusive                      bool                             `json:"project_exclusive"`
	BillingSourceIds                      []int                            `json:"billing_source_ids"`
	BillingSourceExclusive                bool                             `json:"billing_source_exclusive"`
	FundingSourceIds                      []int                            `json:"funding_source_ids"`
	FundingSourceExclusive                bool                             `json:"funding_source_exclusive"`
	CloudProviderIds                      []int                            `json:"cloud_provider_ids"`
	CloudProviderExclusive                bool                             `json:"cloud_provider_exclusive"`
	AccountIds                            []int                            `json:"account_ids"`
	AccountExclusive                      bool                             `json:"account_exclusive"`
	RegionIds                             []int                            `json:"region_ids"`
	RegionExclusive                       bool                             `json:"region_exclusive"`
	ServiceIds                            []int                            `json:"service_ids"`
	ServiceExclusive                      bool                             `json:"service_exclusive"`
	IncludeAppLabelIds                    map[string]interface{}           `json:"include_app_label_ids"`
	ExcludeAppLabelIds                    map[string]interface{}           `json:"exclude_app_label_ids"`
	AppLabelKeyIdDimension                int                              `json:"app_label_key_id_dimension"`
	AppLabelIdsDimension                  []int                            `json:"app_label_ids_dimension"`
	IncludeCloudProviderTagIds            map[string]interface{}           `json:"include_cloud_provider_tag_ids"`
	ExcludeCloudProviderTagIds            map[string]interface{}           `json:"exclude_cloud_provider_tag_ids"`
	CloudProviderTagKeyIdDimension        int                              `json:"cloud_provider_tag_key_id_dimension"`
	CloudProviderTagValueIdsDimension     []int                            `json:"cloud_provider_tag_value_ids_dimension"`
	DeductCredits                         bool                             `json:"deduct_credits"`
	DeductRefunds                         bool                             `json:"deduct_refunds"`
	Scheduled                             bool                             `json:"scheduled"`
	ScheduledFileTypes                    []int                            `json:"scheduled_file_types"`
	ScheduledFileOrientation              int                              `json:"scheduled_file_orientation"`
	ScheduledEmailSubject                 string                           `json:"scheduled_email_subject"`
	ScheduledEmailMessage                 string                           `json:"scheduled_email_message"`
	CreatedAt                             *TimeField                       `json:"created_at"`
	UpdatedAt                             *TimeField                       `json:"updated_at"`
	SavedReportScheduledFrequency         *SpendReportScheduledFrequency   `json:"saved_report_scheduled_frequency"`
}

// SpendReportScheduledFrequency represents the scheduling configuration
type SpendReportScheduledFrequency struct {
	ID                   int        `json:"id,omitempty"`
	Type                 int        `json:"type"`
	DaysOfWeek           []int      `json:"days_of_week"`
	DaysOfMonth          []int      `json:"days_of_month,omitempty"`
	WeeklyRecurrence     *int       `json:"weekly_recurrence"`
	QuarterlyRecurrence  *int       `json:"quarterly_recurrence"`
	Hour                 int        `json:"hour"`
	Minute               int        `json:"minute"`
	TimeZoneIdentifier   string     `json:"time_zone_identifier"`
	StartDate            *TimeField `json:"start_date"`
	EndDate              *TimeField `json:"end_date"`
}

// SpendReportTimeframe represents the timeframe configuration
type SpendReportTimeframe struct {
	Label           string `json:"label"`
	UserSettingKey  string `json:"user_setting_key"`
	StartDatecode   int    `json:"start_datecode"`
	EndDatecode     int    `json:"end_datecode"`
	StartDate       string `json:"start_date"`
	EndDate         string `json:"end_date"`
}

// SpendReportExternalEmail represents external email recipients
type SpendReportExternalEmail struct {
	ID               int        `json:"id,omitempty"`
	SavedReportId    int        `json:"saved_report_id,omitempty"`
	EmailAddress     string     `json:"email_address"`
	UnsubscribedAt   *TimeField `json:"unsubscribed_at,omitempty"`
}

// SpendReportCreateRequest represents the request body for creating a spend report
type SpendReportCreateRequest struct {
	SavedReport      SpendReportCreate           `json:"saved_report"`
	UserIds          []int                       `json:"user_ids"`
	UserGroupIds     []int                       `json:"ugroup_ids"`
	ExcludedUsers    []int                       `json:"excluded_users"`
	ExcludedUgroups  []int                       `json:"excluded_ugroups"`
	ExternalEmails   []SpendReportExternalEmail  `json:"external_emails"`
}

// SpendReportCreate represents the spend report data for creation
type SpendReportCreate struct {
	CreatedBy                             int                              `json:"created_by"`
	ReportName                            string                           `json:"report_name"`
	GlobalVisibility                      bool                             `json:"global_visibility"`
	DateRange                             string                           `json:"date_range"`
	Scope                                 string                           `json:"scope"`
	ScopeId                               int                              `json:"scope_id"`
	StartMonth                            int                              `json:"start_month,omitempty"`
	EndMonth                              int                              `json:"end_month,omitempty"`
	StartDate                             string                           `json:"start_date,omitempty"`
	EndDate                               string                           `json:"end_date,omitempty"`
	Dimension                             string                           `json:"dimension"`
	OUIds                                 []int                            `json:"ou_ids"`
	OUExclusive                           bool                             `json:"ou_exclusive"`
	IncludeDescendants                    bool                             `json:"include_descendants"`
	ProjectIds                            []int                            `json:"project_ids"`
	ProjectExclusive                      bool                             `json:"project_exclusive"`
	BillingSourceIds                      []int                            `json:"billing_source_ids"`
	BillingSourceExclusive                bool                             `json:"billing_source_exclusive"`
	FundingSourceIds                      []int                            `json:"funding_source_ids"`
	FundingSourceExclusive                bool                             `json:"funding_source_exclusive"`
	CloudProviderIds                      []int                            `json:"cloud_provider_ids"`
	CloudProviderExclusive                bool                             `json:"cloud_provider_exclusive"`
	AccountIds                            []int                            `json:"account_ids"`
	AccountExclusive                      bool                             `json:"account_exclusive"`
	ServiceIds                            []int                            `json:"service_ids"`
	ServiceExclusive                      bool                             `json:"service_exclusive"`
	RegionIds                             []int                            `json:"region_ids"`
	RegionExclusive                       bool                             `json:"region_exclusive"`
	IncludeAppLabelIds                    map[string]interface{}           `json:"include_app_label_ids"`
	ExcludeAppLabelIds                    map[string]interface{}           `json:"exclude_app_label_ids"`
	IncludeCloudProviderTagIds            map[string]interface{}           `json:"include_cloud_provider_tag_ids"`
	ExcludeCloudProviderTagIds            map[string]interface{}           `json:"exclude_cloud_provider_tag_ids"`
	Timeframe                             *SpendReportTimeframe            `json:"timeframe,omitempty"`
	TimeGranularityId                     int                              `json:"time_granularity_id"`
	SpendType                             string                           `json:"spend_type"`
	DefaultSpendType                      string                           `json:"default_spend_type,omitempty"`
	DeductCredits                         bool                             `json:"deduct_credits"`
	DeductRefunds                         bool                             `json:"deduct_refunds"`
	Scheduled                             bool                             `json:"scheduled"`
	SavedReportScheduledFrequency         *SpendReportScheduledFrequency   `json:"saved_report_scheduled_frequency,omitempty"`
	ScheduledEmailSubject                 string                           `json:"scheduled_email_subject,omitempty"`
	ScheduledEmailMessage                 string                           `json:"scheduled_email_message,omitempty"`
	ScheduledFileTypes                    []int                            `json:"scheduled_file_types,omitempty"`
	ScheduledFileOrientation              int                              `json:"scheduled_file_orientation,omitempty"`
}

// SpendReportResponse represents the response from creating or getting a spend report
type SpendReportResponse struct {
	Status int `json:"status"`
	Data   struct {
		SavedReport      SpendReport                  `json:"saved_report"`
		ExternalEmails   []SpendReportExternalEmail   `json:"external_emails"`
		UserIds          []int                        `json:"user_ids"`
		UserGroupIds     []int                        `json:"ugroup_ids"`
		Hidden           bool                         `json:"hidden"`
	} `json:"data"`
}

// SpendReportUpdateRequest represents the request body for updating a spend report
type SpendReportUpdateRequest struct {
	SavedReport      SpendReportUpdate           `json:"saved_report"`
	UserIds          []int                       `json:"user_ids"`
	UserGroupIds     []int                       `json:"ugroup_ids"`
	ExcludedUsers    []int                       `json:"excluded_users"`
	ExcludedUgroups  []int                       `json:"excluded_ugroups"`
	ExternalEmails   []SpendReportExternalEmail  `json:"external_emails"`
}

// SpendReportUpdate represents the spend report data for updates
type SpendReportUpdate struct {
	ReportName                            string                           `json:"report_name,omitempty"`
	GlobalVisibility                      *bool                            `json:"global_visibility,omitempty"`
	DateRange                             string                           `json:"date_range,omitempty"`
	Scope                                 string                           `json:"scope,omitempty"`
	ScopeId                               *int                             `json:"scope_id,omitempty"`
	StartMonth                            *int                             `json:"start_month,omitempty"`
	EndMonth                              *int                             `json:"end_month,omitempty"`
	StartDate                             string                           `json:"start_date,omitempty"`
	EndDate                               string                           `json:"end_date,omitempty"`
	Dimension                             string                           `json:"dimension,omitempty"`
	OUIds                                 *[]int                           `json:"ou_ids,omitempty"`
	OUExclusive                           *bool                            `json:"ou_exclusive,omitempty"`
	IncludeDescendants                    *bool                            `json:"include_descendants,omitempty"`
	ProjectIds                            *[]int                           `json:"project_ids,omitempty"`
	ProjectExclusive                      *bool                            `json:"project_exclusive,omitempty"`
	BillingSourceIds                      *[]int                           `json:"billing_source_ids,omitempty"`
	BillingSourceExclusive                *bool                            `json:"billing_source_exclusive,omitempty"`
	FundingSourceIds                      *[]int                           `json:"funding_source_ids,omitempty"`
	FundingSourceExclusive                *bool                            `json:"funding_source_exclusive,omitempty"`
	CloudProviderIds                      *[]int                           `json:"cloud_provider_ids,omitempty"`
	CloudProviderExclusive                *bool                            `json:"cloud_provider_exclusive,omitempty"`
	AccountIds                            *[]int                           `json:"account_ids,omitempty"`
	AccountExclusive                      *bool                            `json:"account_exclusive,omitempty"`
	ServiceIds                            *[]int                           `json:"service_ids,omitempty"`
	ServiceExclusive                      *bool                            `json:"service_exclusive,omitempty"`
	RegionIds                             *[]int                           `json:"region_ids,omitempty"`
	RegionExclusive                       *bool                            `json:"region_exclusive,omitempty"`
	IncludeAppLabelIds                    map[string]interface{}           `json:"include_app_label_ids,omitempty"`
	ExcludeAppLabelIds                    map[string]interface{}           `json:"exclude_app_label_ids,omitempty"`
	IncludeCloudProviderTagIds            map[string]interface{}           `json:"include_cloud_provider_tag_ids,omitempty"`
	ExcludeCloudProviderTagIds            map[string]interface{}           `json:"exclude_cloud_provider_tag_ids,omitempty"`
	Timeframe                             *SpendReportTimeframe            `json:"timeframe,omitempty"`
	TimeGranularityId                     *int                             `json:"time_granularity_id,omitempty"`
	SpendType                             string                           `json:"spend_type,omitempty"`
	DeductCredits                         *bool                            `json:"deduct_credits,omitempty"`
	DeductRefunds                         *bool                            `json:"deduct_refunds,omitempty"`
	Scheduled                             *bool                            `json:"scheduled,omitempty"`
	SavedReportScheduledFrequency         *SpendReportScheduledFrequency   `json:"saved_report_scheduled_frequency,omitempty"`
	ScheduledEmailSubject                 string                           `json:"scheduled_email_subject,omitempty"`
	ScheduledEmailMessage                 string                           `json:"scheduled_email_message,omitempty"`
	ScheduledFileTypes                    *[]int                           `json:"scheduled_file_types,omitempty"`
	ScheduledFileOrientation              *int                             `json:"scheduled_file_orientation,omitempty"`
}