package kionclient

// ProjectEnforcementDetails is the struct for each enforcement detail.
type ProjectEnforcementDetails struct {
	ID                    uint                                `json:"id"`
	Description           string                              `json:"description"`
	Timeframe             string                              `json:"timeframe"`
	SpendOption           string                              `json:"spend_option,omitempty"`
	AmountType            string                              `json:"amount_type,omitempty"`
	Service               *ProjectEnforcementServiceDetails   `json:"service,omitempty"`
	ThresholdType         string                              `json:"threshold_type,omitempty"`
	Threshold             int                                 `json:"threshold"`
	CloudRule             *ProjectEnforcementCloudRuleDetails `json:"cloud_rule,omitempty"`
	Overburn              *bool                               `json:"overburn,omitempty"`
	NotificationFrequency string                              `json:"notification_frequency"`
	ProjectID             int                                 `json:"funding_source_id"`
	Enabled               *bool                               `json:"enabled,omitempty"`
	UserGroupIds          *[]int                              `json:"user_group_ids,omitempty"`
	UserIds               *[]int                              `json:"user_ids,omitempty"`
	Triggered             bool                                `json:"triggered"`
}

// ProjectEnforcementResponse for: GET /api/v3/project/{id}/enforcement
type ProjectEnforcementResponse struct {
	Data   []ProjectEnforcementDetails `json:"data"`
	Status int                         `json:"status"`
}

// ProjectEnforcementCreate for: POST /api/v3/project/{id}/enforcement
type ProjectEnforcementCreate struct {
	ID            int    `json:"id"`
	Description   string `json:"description"`
	Timeframe     string `json:"timeframe"`
	SpendOption   string `json:"spend_option,omitempty"`
	AmountType    string `json:"amount_type,omitempty"`
	ServiceID     *int   `json:"service_id,omitempty"`
	ThresholdType string `json:"threshold_type,omitempty"`
	Threshold     int    `json:"threshold"`
	CloudRuleID   *int   `json:"cloud_rule_id,omitempty"`
	Overburn      *bool  `json:"overburn,omitempty"`
	UserGroupIds  *[]int `json:"user_group_ids,omitempty"`
	UserIds       *[]int `json:"user_ids,omitempty"`
}

// ProjectEnforcementUpdate for: PATCH /api/v3/project/{id}/enforcement/{enforcement_id}
type ProjectEnforcementUpdate struct {
	Description   string `json:"description"`
	Timeframe     string `json:"timeframe"`
	SpendOption   string `json:"spend_option,omitempty"`
	AmountType    string `json:"amount_type,omitempty"`
	ServiceID     *int   `json:"service_id,omitempty"`
	ThresholdType string `json:"threshold_type,omitempty"`
	Threshold     int    `json:"threshold"`
	CloudRuleID   *int   `json:"cloud_rule_id,omitempty"`
	Overburn      *bool  `json:"overburn,omitempty"`
	Enabled       *bool  `json:"enabled,omitempty"`
	UserGroupIds  *[]int `json:"user_group_ids,omitempty"`
	UserIds       *[]int `json:"user_ids,omitempty"`
}

// ProjectEnforcementUsersCreate for: POST /api/v3/project/{id}/enforcement/{enforcement_id}/user
type ProjectEnforcementUsersCreate struct {
	UserGroupIds *[]int `json:"user_group_ids"`
	UserIds      *[]int `json:"user_ids"`
}

type ProjectEnforcementCloudRuleDetails struct {
	BuiltIn       bool   `json:"built_in"`
	Description   string `json:"description"`
	ID            int    `json:"id"`
	Name          string `json:"name"`
	PostWebhookID uint   `json:"post_webhook_id"`
	PreWebhookID  uint   `json:"pre_webhook_id"`
}

type ProjectEnforcementServiceDetails struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProviderType string `json:"provider_type"`
}
