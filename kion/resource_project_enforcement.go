package kion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceProjectEnforcement() *schema.Resource {
	return &schema.Resource{
		Description: "Manages enforcement rules for projects to control service usage based on various criteria like" +
			"spend limits and timeframe restrictions. .\n\n" +
			"This resource allows for creating, reading, updating, and deleting project-specific enforcement settings.",
		CreateContext: resourceProjectEnforcementCreate,
		ReadContext:   resourceProjectEnforcementRead,
		UpdateContext: resourceProjectEnforcementUpdate,
		DeleteContext: resourceProjectEnforcementDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional, user-provided description of the enforcement.",
			},
			"timeframe": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"lifetime", "month", "year", "funding_source"}, false),
				Description:  "Timeframe of the enforcement. Valid values are 'lifetime', 'month', 'year', 'funding_source'.",
			},
			"spend_option": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"spend", "remaining", "spend_rate"}, false),
				Description:  "Type of spend option. Valid values are 'spend', 'remaining', 'spend_rate'.",
			},
			"amount_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"custom", "last_month"}, false),
				Description:  "Type of the amount. Valid values are 'custom', 'last_month'.",
			},
			"service_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the service related to the enforcement.",
			},
			"threshold_type": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"dollar", "percent"}, false),
				Description:  "Type of the threshold value. Valid values are 'dollar', 'percent'.",
			},
			"threshold": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Threshold value. Either a dollar amount or a percentage, depending on the threshold type.",
			},
			"cloud_rule_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Defines a Cloud Rule ID associated with the enforcement.",
			},
			"notification_frequency": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Frequency at which notifications are sent for this enforcement.",
			},
			"project_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "ID of the project under enforcement.",
			},
			"ou_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "If enforcement is from an Organizational Unit (OU), this is the ID of the OU.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Flag that specifies if the enforcement is enabled.",
			},
			"overburn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Flag that specifies if enforcement will place project in an overburn state when triggered.",
			},
			"user_group_ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Optional:    true,
				Description: "List of user group IDs that will receive notifications from the enforcement.",
			},
			"user_ids": {
				Type:        schema.TypeList,
				Elem:        &schema.Schema{Type: schema.TypeInt},
				Optional:    true,
				Description: "List of user IDs that will receive notifications from the enforcement.",
			},
			"triggered": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Flag that specifies if the enforcement is currently triggered.",
			},
		},
	}
}

func resourceProjectEnforcementCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	projectID, ok := d.GetOk("project_id")
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid or Missing Project ID",
			Detail:   "The project ID is either missing or not valid. Please provide a valid project ID.",
		})
		return diags
	}
	projectIDInt := projectID.(int)

	post := hc.ProjectEnforcementCreate{
		Description:   d.Get("description").(string),
		Timeframe:     d.Get("timeframe").(string),
		SpendOption:   d.Get("spend_option").(string),
		AmountType:    d.Get("amount_type").(string),
		ServiceID:     hc.OptionalInt(d, "service_id"),
		ThresholdType: d.Get("threshold_type").(string),
		Threshold:     d.Get("threshold").(int),
		CloudRuleID:   hc.OptionalInt(d, "cloud_rule_id"),
		Overburn:      d.Get("overburn").(bool),
		UserGroupIds:  convertToIntSlice(d.Get("user_group_ids").([]interface{})),
		UserIds:       convertToIntSlice(d.Get("user_ids").([]interface{})),
	}

	// Build the endpoint URL for creation
	projectEnforcementURL := fmt.Sprintf("/v3/project/%d/enforcement", projectIDInt)

	// Send the create request
	if rb, err := json.Marshal(post); err == nil {
		tflog.Debug(ctx, fmt.Sprintf("Creating Project Enforcement via POST %s", projectEnforcementURL), map[string]interface{}{"postData": string(rb)})
	}

	resp, err := client.POST(projectEnforcementURL, post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to Create Project Enforcement",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed Project Enforcement Creation",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return diags
}

func resourceProjectEnforcementRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	enforcementID := d.Id()
	projectID, ok := d.GetOk("project_id")
	if !ok {
		return diag.Errorf("Invalid or missing project ID")
	}

	enforcementIDInt, err := strconv.Atoi(enforcementID)
	if err != nil {
		return diag.FromErr(err)
	}

	resp := new(hc.ProjectEnforcementResponse)
	err = client.GET(fmt.Sprintf("/v3/project/%d/enforcement", projectID), resp)
	if err != nil {
		return diag.FromErr(err)
	}

	if len(resp.Data) == 0 {
		return diag.Errorf("No Project Enforcement data found for Project ID %d", projectID)
	}

	var found bool
	for _, item := range resp.Data {
		if int(item.ID) == enforcementIDInt {
			diags = append(diags, safeSet(d, "description", item.Description)...)
			diags = append(diags, safeSet(d, "timeframe", item.Timeframe)...)
			diags = append(diags, safeSet(d, "spend_option", item.SpendOption)...)
			diags = append(diags, safeSet(d, "amount_type", item.AmountType)...)
			diags = append(diags, safeSet(d, "threshold_type", item.ThresholdType)...)
			diags = append(diags, safeSet(d, "threshold", item.Threshold)...)
			diags = append(diags, safeSet(d, "enabled", item.Enabled)...)
			diags = append(diags, safeSet(d, "overburn", item.Overburn)...)
			diags = append(diags, safeSet(d, "user_group_ids", item.UserGroupIds)...)
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("Enforcement ID %d not found under Project ID %d", enforcementIDInt, projectID)
	}

	return diags
}

func resourceProjectEnforcementUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	projectID, ok := d.GetOk("project_id")
	if !ok {
		return diag.Errorf("Invalid or missing project ID")
	}

	projectIDInt := projectID.(int)
	enforcementID := d.Id()

	// Prepare the update request
	req := hc.ProjectEnforcementUpdate{
		Description:   d.Get("description").(string),
		Timeframe:     d.Get("timeframe").(string),
		SpendOption:   d.Get("spend_option").(string),
		AmountType:    d.Get("amount_type").(string),
		ServiceID:     hc.OptionalInt(d, "service_id"),
		ThresholdType: d.Get("threshold_type").(string),
		Threshold:     d.Get("threshold").(int),
		CloudRuleID:   hc.OptionalInt(d, "cloud_rule_id"),
		Overburn:      d.Get("overburn").(bool),
		Enabled:       d.Get("enabled").(bool),
	}

	// Send the update request
	endpoint := fmt.Sprintf("/v3/project/%d/enforcement/%s", projectIDInt, enforcementID)
	err := client.PATCH(endpoint, req)
	if err != nil {
		return diag.Errorf("Unable to update Project Enforcement: %v", err)
	}

	// Always perform a read after update to ensure state is consistent
	readDiags := resourceProjectEnforcementRead(ctx, d, m)
	diags = append(diags, readDiags...)
	return diags
}

func resourceProjectEnforcementDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	// Fetching projectID safely
	projectID, ok := d.GetOk("project_id")
	if !ok {
		return diag.Errorf("Invalid or missing project ID")
	}

	// Converting projectID to int, assuming it's stored as an integer in the state
	projectIDInt, ok := projectID.(int)
	if !ok {
		return diag.Errorf("Project ID should be an integer")
	}

	enforcementID := d.Id()
	if enforcementID == "" {
		return diag.Errorf("Invalid or missing enforcement ID")
	}

	// Preparing the endpoint URL
	endpoint := fmt.Sprintf("/v3/project/%d/enforcement/%s", projectIDInt, enforcementID)
	err := client.DELETE(endpoint, nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Project Enforcement",
			Detail:   fmt.Sprintf("Error: %v when attempting to delete the enforcement with ID: %s", err.Error(), enforcementID),
		})
		return diags
	}

	// Clear the ID from the Terraform state as the resource no longer exists
	d.SetId("")

	return diags
}

// Helper function to convert []interface{} from Terraform state to []int required by the Kion API.
func convertToIntSlice(interfaceSlice []interface{}) []int {
	intSlice := make([]int, len(interfaceSlice))
	for i, v := range interfaceSlice {
		intSlice[i] = v.(int)
	}
	return intSlice
}

// Handle setting Terraform schema values, centralizing error reporting
func safeSet(d *schema.ResourceData, key string, value interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	if err := d.Set(key, value); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error setting field",
			Detail:   fmt.Sprintf("Error setting %s: %s", key, err),
		})
	}
	return diags
}
