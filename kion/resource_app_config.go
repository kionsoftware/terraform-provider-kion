package kion

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceAppConfig() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAppConfigUpdate, // No separate create, just update
		ReadContext:   resourceAppConfigRead,
		UpdateContext: resourceAppConfigUpdate,
		DeleteContext: resourceAppConfigDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				// Set a static ID for app-config since there's only one
				d.SetId("app-config")
				resourceAppConfigRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"all_users_see_ou_names": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether all users can see the names of OU's in the organization chart.",
			},
			"allocation_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if allocation mode is enabled in the application.",
			},
			"allow_custom_permission_schemes": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether custom permission schemes are allowed or not.",
			},
			"app_api_key_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if App API Key creation is enabled.",
			},
			"app_api_key_lifespan": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Indicates the lifespan of App API Keys in days.",
			},
			"app_api_key_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Indicates the max amount of App API Keys per user.",
			},
			"aws_access_key_creation_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether AWS access keys creation is enabled.",
			},
			"budget_mode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if budget mode is enabled in the application.",
			},
			"cloud_rule_group_ownership_only": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if cloud rules are restricted to User Group ownership only. Setting this to true will remove all users from cloud rules. This cannot be undone.",
			},
			"cost_savings_allow_terminate": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether resource termination is allowed in-app.",
			},
			"cost_savings_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether Cost Savings is enabled or not.",
			},
			"cost_savings_post_token_life_hours": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Post token life (hours) for Cloud Custodian webhook actions to execute.",
			},
			"default_org_chart_view": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Defines the default organization chart view.",
				ValidateFunc: validateOrgChartView,
			},
			"enforce_funding": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether spend plans or budgets must be created on all projects.",
			},
			"enforce_funding_sources": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether every project should have a funding source.",
			},
			"event_driven_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether event driven is enabled or not.",
			},
			"forecasting_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether forecasting is enabled or not.",
			},
			"idms_groups_as_viewers_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether, when new groups are created via IDMS, the group will be set as a viewer by default.",
			},
			"reserved_instances_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether reserved instances are enabled or not.",
			},
			"resource_inventory_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether resource inventory is enabled or not.",
			},
			"saml_debug": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether SAML debugging is enabled or not.",
			},
			"smtp_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates whether SMTP is enabled or not.",
			},
			"smtp_from": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SMTP from address.",
			},
			"smtp_host": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SMTP host.",
			},
			"smtp_password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The SMTP password.",
			},
			"smtp_port": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The SMTP port.",
			},
			"smtp_skip_verify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the app should skip SMTP verification.",
			},
			"smtp_username": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The SMTP username.",
			},
			"supported_aws_regions": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of supported AWS regions.",
			},
		},
	}
}

func validateOrgChartView(val interface{}, key string) (warns []string, errs []error) {
	v := val.(string)
	validValues := []string{"list", "policy", "compliance", "financial", "spend"}

	for _, valid := range validValues {
		if v == valid {
			return
		}
	}

	errs = append(errs, fmt.Errorf("%q must be one of %v, got: %q", key, validValues, v))
	return
}

func resourceAppConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
	d.SetId("app-config")

	return diags
}

func resourceAppConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	appConfig := &hc.AppConfig{}

	// Use helper functions to populate only configured fields
	if boolPtr := hc.OptionalValue[bool](d, "all_users_see_ou_names"); boolPtr != nil {
		appConfig.AllUsersSeeOUNames = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "allocation_mode"); boolPtr != nil {
		appConfig.AllocationMode = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "allow_custom_permission_schemes"); boolPtr != nil {
		appConfig.AllowCustomPermissionSchemes = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "app_api_key_creation_enabled"); boolPtr != nil {
		appConfig.AppAPIKeyCreationEnabled = *boolPtr
	}
	if intPtr := hc.OptionalValue[int](d, "app_api_key_lifespan"); intPtr != nil {
		appConfig.AppAPIKeyLifespan = int64(*intPtr)
	}
	if intPtr := hc.OptionalValue[int](d, "app_api_key_limit"); intPtr != nil {
		appConfig.AppAPIKeyLimit = int64(*intPtr)
	}
	if boolPtr := hc.OptionalValue[bool](d, "aws_access_key_creation_enabled"); boolPtr != nil {
		appConfig.AWSAccessKeyCreationEnabled = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "budget_mode"); boolPtr != nil {
		appConfig.BudgetMode = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "cloud_rule_group_ownership_only"); boolPtr != nil {
		appConfig.CloudRuleGroupOwnershipOnly = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "cost_savings_allow_terminate"); boolPtr != nil {
		appConfig.CostSavingsAllowTerminate = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "cost_savings_enabled"); boolPtr != nil {
		appConfig.CostSavingsEnabled = *boolPtr
	}
	if intPtr := hc.OptionalValue[int](d, "cost_savings_post_token_life_hours"); intPtr != nil {
		appConfig.CostSavingsPostTokenLifeHours = uint64(*intPtr)
	}
	if strPtr := hc.OptionalValue[string](d, "default_org_chart_view"); strPtr != nil {
		appConfig.DefaultOrgChartView = *strPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "enforce_funding"); boolPtr != nil {
		appConfig.EnforceFunding = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "enforce_funding_sources"); boolPtr != nil {
		appConfig.EnforceFundingSources = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "event_driven_enabled"); boolPtr != nil {
		appConfig.EventDrivenEnabled = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "forecasting_enabled"); boolPtr != nil {
		appConfig.ForecastingEnabled = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "idms_groups_as_viewers_default"); boolPtr != nil {
		appConfig.IDMSGroupsAsViewersDefault = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "reserved_instances_enabled"); boolPtr != nil {
		appConfig.ReservedInstancesEnabled = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "resource_inventory_enabled"); boolPtr != nil {
		appConfig.ResourceInventoryEnabled = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "saml_debug"); boolPtr != nil {
		appConfig.SAMLDebug = *boolPtr
	}
	if boolPtr := hc.OptionalValue[bool](d, "smtp_enabled"); boolPtr != nil {
		appConfig.SMTPEnabled = *boolPtr
	}
	if strPtr := hc.OptionalValue[string](d, "smtp_from"); strPtr != nil {
		appConfig.SMTPFrom = *strPtr
	}
	if strPtr := hc.OptionalValue[string](d, "smtp_host"); strPtr != nil {
		appConfig.SMTPHost = *strPtr
	}
	if strPtr := hc.OptionalValue[string](d, "smtp_password"); strPtr != nil {
		appConfig.SMTPPassword = *strPtr
	}
	if intPtr := hc.OptionalValue[int](d, "smtp_port"); intPtr != nil {
		appConfig.SMTPPort = int64(*intPtr)
	}
	if boolPtr := hc.OptionalValue[bool](d, "smtp_skip_verify"); boolPtr != nil {
		appConfig.SMTPSkipVerify = *boolPtr
	}
	if strPtr := hc.OptionalValue[string](d, "smtp_username"); strPtr != nil {
		appConfig.SMTPUsername = *strPtr
	}
	if v, ok := d.GetOk("supported_aws_regions"); ok {
		appConfig.SupportedAWSRegions = hc.FlattenStringArray(v.([]interface{}))
	}

	err := client.PATCH("/v3/app-config", appConfig)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("app-config")
	return resourceAppConfigRead(ctx, d, m)
}

func resourceAppConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// App config cannot be deleted, just remove from Terraform state
	d.SetId("")
	return nil
}
