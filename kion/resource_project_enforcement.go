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
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"spend", "remaining"}, false),
				Description:  "Type of spend option. Valid values are 'spend', 'remaining'.",
			},
			"amount_type": {
				Type:         schema.TypeString,
				Optional:     true,
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
				Optional:     true,
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
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Flag that specifies if the enforcement is enabled.",
			},
			"overburn": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag that specifies if enforcement will place project in an overburn state when triggered.",
			},
			"user_group_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeInt},
				Optional:     true,
				Description:  "List of user group IDs that will receive notifications from the enforcement.",
				AtLeastOneOf: []string{"user_group_ids", "user_ids"},
			},
			"user_ids": {
				Type:         schema.TypeList,
				Elem:         &schema.Schema{Type: schema.TypeInt},
				Optional:     true,
				Description:  "List of user IDs that will receive notifications from the enforcement.",
				AtLeastOneOf: []string{"user_group_ids", "user_ids"},
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

	userGroupIds := hc.FlattenGenericIDPointer(d, "user_group_ids")
	userIds := hc.FlattenGenericIDPointer(d, "user_ids")

	post := hc.ProjectEnforcementCreate{
		Description:   d.Get("description").(string),
		Timeframe:     d.Get("timeframe").(string),
		SpendOption:   d.Get("spend_option").(string),
		AmountType:    d.Get("amount_type").(string),
		ServiceID:     hc.OptionalInt(d, "service_id"),
		ThresholdType: d.Get("threshold_type").(string),
		Threshold:     d.Get("threshold").(int),
		CloudRuleID:   hc.OptionalInt(d, "cloud_rule_id"),
		Overburn:      hc.OptionalBool(d, "overburn"),
		UserGroupIds:  userGroupIds,
		UserIds:       userIds,
	}

	// Ensure at least one user group or user is provided
	if (userGroupIds == nil || len(*userGroupIds) == 0) && (userIds == nil || len(*userIds) == 0) {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Invalid User or User Group",
			Detail:   "At least one user or user group must be specified.",
		})
		return diags
	}

	// Debugging: Log the post payload
	if rb, err := json.Marshal(post); err == nil {
		tflog.Debug(ctx, fmt.Sprintf("Creating Project Enforcement with payload: %s", string(rb)))
	}

	// Build the endpoint URL for creation
	projectEnforcementURL := fmt.Sprintf("/v3/project/%d/enforcement", projectIDInt)

	// Send the create request
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
			diags = append(diags, safeSet(d, "user_ids", item.UserIds)...)
			found = true
			break
		}
	}

	if !found {
		return diag.Errorf("Enforcement ID %d not found under Project ID %d", enforcementIDInt, projectID)
	}

	return diags
}

func AddProjectEnforcementUsers(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	enforcementID := d.Id()
	projectID, ok := d.GetOk("project_id")
	if !ok {
		return diag.Errorf("Invalid or missing project ID")
	}

	projectIDInt, ok := projectID.(int)
	if !ok {
		return diag.Errorf("Project ID should be an integer, got %T", projectID)
	}

	arrAddUserIds := hc.FlattenGenericIDPointer(d, "user_ids")
	arrAddUserGroupIds := hc.FlattenGenericIDPointer(d, "user_group_ids")

	if arrAddUserIds == nil && arrAddUserGroupIds == nil {
		return diag.Errorf("At least one user ID or user group ID must be provided")
	}

	req := hc.ProjectEnforcementUsersCreate{
		UserIds:      arrAddUserIds,
		UserGroupIds: arrAddUserGroupIds,
	}

	endpoint := fmt.Sprintf("/v3/project/%d/enforcement/%s/user", projectIDInt, enforcementID)
	_, err := client.POST(endpoint, req)
	if err != nil {
		return diag.Errorf("Error adding users/user groups in Project Enforcement: %v", err)
	}

	return diags
}

func RemoveProjectEnforcementUsers(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	enforcementID := d.Id()
	projectID, ok := d.GetOk("project_id")
	if !ok {
		return diag.Errorf("Invalid or missing project ID")
	}

	projectIDInt, ok := projectID.(int)
	if !ok {
		return diag.Errorf("Project ID should be an integer, got %T", projectID)
	}

	// Get the current state of users and user groups
	currentUserIds := hc.FlattenGenericIDPointer(d, "user_ids")
	currentUserGroupIds := hc.FlattenGenericIDPointer(d, "user_group_ids")

	// Get the previous state to identify what needs to be removed
	prevUserIds, prevUserGroupIds := getPreviousEnforcementUserAndGroupIds(d)

	// Determine which user IDs and user group IDs need to be removed
	toRemoveUserIds := findEnforcementIdDifferences(prevUserIds, *currentUserIds)
	toRemoveUserGroupIds := findEnforcementIdDifferences(prevUserGroupIds, *currentUserGroupIds)

	// If there's nothing to remove, return early
	if len(toRemoveUserIds) == 0 && len(toRemoveUserGroupIds) == 0 {
		return diags
	}

	req := hc.ProjectEnforcementUsersCreate{
		UserIds:      &toRemoveUserIds,
		UserGroupIds: &toRemoveUserGroupIds,
	}

	endpoint := fmt.Sprintf("/v3/project/%d/enforcement/%s/user", projectIDInt, enforcementID)
	err := client.DELETE(endpoint, req)
	if err != nil {
		return diag.Errorf("Error removing users/user groups in Project Enforcement: %v", err)
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

	projectIDInt, ok := projectID.(int)
	if !ok {
		return diag.Errorf("Project ID should be an integer")
	}

	enforcementID := d.Id()

	// Determine if the attributes that are updatable have changed.
	if d.HasChanges("description", "timeframe", "spend_option", "amount_type", "service_id", "threshold_type", "threshold", "cloud_rule_id", "overburn", "enabled") {
		req := hc.ProjectEnforcementUpdate{
			Description:   d.Get("description").(string),
			Timeframe:     d.Get("timeframe").(string),
			SpendOption:   d.Get("spend_option").(string),
			AmountType:    d.Get("amount_type").(string),
			ServiceID:     hc.OptionalInt(d, "service_id"),
			ThresholdType: d.Get("threshold_type").(string),
			Threshold:     d.Get("threshold").(int),
			CloudRuleID:   hc.OptionalInt(d, "cloud_rule_id"),
			Overburn:      hc.OptionalBool(d, "overburn"),
			Enabled:       hc.OptionalBool(d, "enabled"),
		}

		// Send the update request
		endpoint := fmt.Sprintf("/v3/project/%d/enforcement/%s", projectIDInt, enforcementID)
		err := client.PATCH(endpoint, req)
		if err != nil {
			return diag.Errorf("Unable to update Project Enforcement: %v", err)
		}
	}

	// Check for changes in user groups or users
	if d.HasChange("user_group_ids") || d.HasChange("user_ids") {
		// First add the new users/user groups to ensure there is always at least one
		addDiags := AddProjectEnforcementUsers(ctx, d, m)
		diags = append(diags, addDiags...)

		// Then remove any existing users/user groups that are no longer needed
		removeDiags := RemoveProjectEnforcementUsers(ctx, d, m)
		diags = append(diags, removeDiags...)
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

	// Converting projectID to int
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

// safeSet handles setting Terraform schema values, centralizing error reporting and ensuring non-nil values.
func safeSet(d *schema.ResourceData, key string, value interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	// Check if the value is non-nil before setting it in the schema
	if value != nil {
		// Attempt to set the value in the schema
		if err := d.Set(key, value); err != nil {
			// Append a diagnostic message if there's an error while setting the value
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Error setting field",
				Detail:   fmt.Sprintf("Error setting %s: %s", key, err),
			})
		}
	}
	return diags
}

// getPreviousEnforcementUserAndGroupIds retrieves the previous state of user and user group IDs
// from the Terraform resource data.
func getPreviousEnforcementUserAndGroupIds(d *schema.ResourceData) ([]int, []int) {
	var prevUserIds, prevUserGroupIds []int

	// Check if the "user_ids" field has changed
	if d.HasChange("user_ids") {
		// Get the previous value of the "user_ids" field
		oldValue, _ := d.GetChange("user_ids")
		// Convert the previous value to a slice of integers
		prevUserIds = convertInterfaceSliceToIntSliceEnforcement(oldValue.([]interface{}))
	}

	// Check if the "user_group_ids" field has changed
	if d.HasChange("user_group_ids") {
		// Get the previous value of the "user_group_ids" field
		oldValue, _ := d.GetChange("user_group_ids")
		// Convert the previous value to a slice of integers
		prevUserGroupIds = convertInterfaceSliceToIntSliceEnforcement(oldValue.([]interface{}))
	}

	return prevUserIds, prevUserGroupIds
}

// convertInterfaceSliceToIntSliceEnforcement converts a slice of interfaces to a slice of integers
// for enforcement purposes.
func convertInterfaceSliceToIntSliceEnforcement(interfaceSlice []interface{}) []int {
	// Create a slice of integers with the same length as the input slice
	intSlice := make([]int, len(interfaceSlice))
	// Iterate over the input slice, casting each element to an integer
	for i, v := range interfaceSlice {
		intSlice[i] = v.(int)
	}
	return intSlice
}

// findEnforcementIdDifferences finds the differences between two slices of integers,
// returning the elements that are present in slice1 but not in slice2.
func findEnforcementIdDifferences(slice1, slice2 []int) []int {
	// Create a set from the second slice for efficient lookups
	set := make(map[int]bool)
	for _, v := range slice2 {
		set[v] = true
	}

	var diff []int
	// Iterate over the first slice and find elements not present in the second slice
	for _, v := range slice1 {
		if !set[v] {
			diff = append(diff, v)
		}
	}
	return diff
}
