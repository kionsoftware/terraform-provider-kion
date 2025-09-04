package kion

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceAppConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAppConfigRead,
		Schema: map[string]*schema.Schema{
			"all_users_see_ou_names": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether all users can see the names of OU's in the organization chart.",
			},
			"allocation_mode": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if allocation mode is enabled in the application.",
			},
			"allow_custom_permission_schemes": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether custom permission schemes are allowed or not.",
			},
			"app_api_key_creation_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if App API Key creation is enabled.",
			},
			"app_api_key_lifespan": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates the lifespan of App API Keys in days.",
			},
			"app_api_key_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates the max amount of App API Keys per user.",
			},
			"aws_access_key_creation_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether AWS access keys creation is enabled.",
			},
			"budget_mode": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if budget mode is enabled in the application.",
			},
			"cloud_rule_group_ownership_only": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if cloud rules are restricted to User Group ownership only.",
			},
			"cost_savings_allow_terminate": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether resource termination is allowed in-app.",
			},
			"cost_savings_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether Cost Savings is enabled or not.",
			},
			"cost_savings_post_token_life_hours": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Post token life (hours) for Cloud Custodian webhook actions to execute.",
			},
			"default_org_chart_view": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the default organization chart view.",
			},
			"enforce_funding": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether spend plans or budgets must be created on all projects.",
			},
			"enforce_funding_sources": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether every project should have a funding source.",
			},
			"event_driven_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether event driven is enabled or not.",
			},
			"forecasting_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether forecasting is enabled or not.",
			},
			"idms_groups_as_viewers_default": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether, when new groups are created via IDMS, the group will be set as a viewer by default.",
			},
			"reserved_instances_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether reserved instances are enabled or not.",
			},
			"resource_inventory_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether resource inventory is enabled or not.",
			},
			"saml_debug": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether SAML debugging is enabled or not.",
			},
			"smtp_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether SMTP is enabled or not.",
			},
			"smtp_from": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SMTP from address.",
			},
			"smtp_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SMTP host.",
			},
			"smtp_password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The SMTP password.",
			},
			"smtp_port": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The SMTP port.",
			},
			"smtp_skip_verify": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the app should skip SMTP verification.",
			},
			"smtp_username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The SMTP username.",
			},
			"supported_aws_regions": {
				Type:        schema.TypeList,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of supported AWS regions.",
			},
		},
	}
}

func dataSourceAppConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	var diags diag.Diagnostics

	var reqPayload hc.AppConfigResponse
	err := client.GET("/v3/app-config", &reqPayload)
	if err != nil {
		return diag.FromErr(err)
	}

	if reqPayload.Data == nil {
		return diag.Errorf("app-config data is nil")
	}

	appConfig := reqPayload.Data

	// Use SafeSet for better error handling
	diags = append(diags, hc.SafeSet(d, "all_users_see_ou_names", appConfig.AllUsersSeeOUNames, "Error setting all_users_see_ou_names")...)
	diags = append(diags, hc.SafeSet(d, "allocation_mode", appConfig.AllocationMode, "Error setting allocation_mode")...)
	diags = append(diags, hc.SafeSet(d, "allow_custom_permission_schemes", appConfig.AllowCustomPermissionSchemes, "Error setting allow_custom_permission_schemes")...)
	diags = append(diags, hc.SafeSet(d, "app_api_key_creation_enabled", appConfig.AppAPIKeyCreationEnabled, "Error setting app_api_key_creation_enabled")...)
	diags = append(diags, hc.SafeSet(d, "app_api_key_lifespan", appConfig.AppAPIKeyLifespan, "Error setting app_api_key_lifespan")...)
	diags = append(diags, hc.SafeSet(d, "app_api_key_limit", appConfig.AppAPIKeyLimit, "Error setting app_api_key_limit")...)
	diags = append(diags, hc.SafeSet(d, "aws_access_key_creation_enabled", appConfig.AWSAccessKeyCreationEnabled, "Error setting aws_access_key_creation_enabled")...)
	diags = append(diags, hc.SafeSet(d, "budget_mode", appConfig.BudgetMode, "Error setting budget_mode")...)
	diags = append(diags, hc.SafeSet(d, "cloud_rule_group_ownership_only", appConfig.CloudRuleGroupOwnershipOnly, "Error setting cloud_rule_group_ownership_only")...)
	diags = append(diags, hc.SafeSet(d, "cost_savings_allow_terminate", appConfig.CostSavingsAllowTerminate, "Error setting cost_savings_allow_terminate")...)
	diags = append(diags, hc.SafeSet(d, "cost_savings_enabled", appConfig.CostSavingsEnabled, "Error setting cost_savings_enabled")...)
	diags = append(diags, hc.SafeSet(d, "cost_savings_post_token_life_hours", appConfig.CostSavingsPostTokenLifeHours, "Error setting cost_savings_post_token_life_hours")...)
	diags = append(diags, hc.SafeSet(d, "default_org_chart_view", appConfig.DefaultOrgChartView, "Error setting default_org_chart_view")...)
	diags = append(diags, hc.SafeSet(d, "enforce_funding", appConfig.EnforceFunding, "Error setting enforce_funding")...)
	diags = append(diags, hc.SafeSet(d, "enforce_funding_sources", appConfig.EnforceFundingSources, "Error setting enforce_funding_sources")...)
	diags = append(diags, hc.SafeSet(d, "event_driven_enabled", appConfig.EventDrivenEnabled, "Error setting event_driven_enabled")...)
	diags = append(diags, hc.SafeSet(d, "forecasting_enabled", appConfig.ForecastingEnabled, "Error setting forecasting_enabled")...)
	diags = append(diags, hc.SafeSet(d, "idms_groups_as_viewers_default", appConfig.IDMSGroupsAsViewersDefault, "Error setting idms_groups_as_viewers_default")...)
	diags = append(diags, hc.SafeSet(d, "reserved_instances_enabled", appConfig.ReservedInstancesEnabled, "Error setting reserved_instances_enabled")...)
	diags = append(diags, hc.SafeSet(d, "resource_inventory_enabled", appConfig.ResourceInventoryEnabled, "Error setting resource_inventory_enabled")...)
	diags = append(diags, hc.SafeSet(d, "saml_debug", appConfig.SAMLDebug, "Error setting saml_debug")...)
	diags = append(diags, hc.SafeSet(d, "smtp_enabled", appConfig.SMTPEnabled, "Error setting smtp_enabled")...)
	diags = append(diags, hc.SafeSet(d, "smtp_from", appConfig.SMTPFrom, "Error setting smtp_from")...)
	diags = append(diags, hc.SafeSet(d, "smtp_host", appConfig.SMTPHost, "Error setting smtp_host")...)
	diags = append(diags, hc.SafeSet(d, "smtp_password", appConfig.SMTPPassword, "Error setting smtp_password")...)
	diags = append(diags, hc.SafeSet(d, "smtp_port", appConfig.SMTPPort, "Error setting smtp_port")...)
	diags = append(diags, hc.SafeSet(d, "smtp_skip_verify", appConfig.SMTPSkipVerify, "Error setting smtp_skip_verify")...)
	diags = append(diags, hc.SafeSet(d, "smtp_username", appConfig.SMTPUsername, "Error setting smtp_username")...)
	diags = append(diags, hc.SafeSet(d, "supported_aws_regions", appConfig.SupportedAWSRegions, "Error setting supported_aws_regions")...)

	// Use a static ID since there's only one app config
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
