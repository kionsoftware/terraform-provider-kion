package kionclient

// AppConfig defines an AppConfig.
type AppConfig struct {
	AllUsersSeeOUNames            bool     `json:"all_users_see_ou_names,omitempty"`
	AllocationMode                bool     `json:"allocation_mode,omitempty"`
	AllowCustomPermissionSchemes  bool     `json:"allow_custom_permission_schemes,omitempty"`
	AppAPIKeyCreationEnabled      bool     `json:"app_api_key_creation_enabled,omitempty"`
	AppAPIKeyLifespan             int64    `json:"app_api_key_lifespan,omitempty"`
	AppAPIKeyLimit                int64    `json:"app_api_key_limit,omitempty"`
	AWSAccessKeyCreationEnabled   bool     `json:"aws_access_key_creation_enabled,omitempty"`
	BudgetMode                    bool     `json:"budget_mode,omitempty"`
	CloudRuleGroupOwnershipOnly   bool     `json:"cloud_rule_group_ownership_only,omitempty"`
	CostSavingsAllowTerminate     bool     `json:"cost_savings_allow_terminate,omitempty"`
	CostSavingsEnabled            bool     `json:"cost_savings_enabled,omitempty"`
	CostSavingsPostTokenLifeHours uint64   `json:"cost_savings_post_token_life_hours,omitempty"`
	DefaultOrgChartView           string   `json:"default_org_chart_view,omitempty"`
	EnforceFunding                bool     `json:"enforce_funding,omitempty"`
	EnforceFundingSources         bool     `json:"enforce_funding_sources,omitempty"`
	EventDrivenEnabled            bool     `json:"event_driven_enabled,omitempty"`
	ForecastingEnabled            bool     `json:"forecasting_enabled,omitempty"`
	IDMSGroupsAsViewersDefault    bool     `json:"idms_groups_as_viewers_default,omitempty"`
	ReservedInstancesEnabled      bool     `json:"reserved_instances_enabled,omitempty"`
	ResourceInventoryEnabled      bool     `json:"resource_inventory_enabled,omitempty"`
	SAMLDebug                     bool     `json:"saml_debug,omitempty"`
	SMTPEnabled                   bool     `json:"smtp_enabled,omitempty"`
	SMTPFrom                      string   `json:"smtp_from,omitempty"`
	SMTPHost                      string   `json:"smtp_host,omitempty"`
	SMTPPassword                  string   `json:"smtp_password,omitempty"`
	SMTPPort                      int64    `json:"smtp_port,omitempty"`
	SMTPSkipVerify                bool     `json:"smtp_skip_verify,omitempty"`
	SMTPUsername                  string   `json:"smtp_username,omitempty"`
	SupportedAWSRegions           []string `json:"supported_aws_regions,omitempty"`
}

// AppConfigResponse defines the response for app config operations.
type AppConfigResponse struct {
	Data   *AppConfig `json:"data,omitempty"`
	Status int        `json:"status,omitempty"`
}
