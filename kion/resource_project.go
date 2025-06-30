package kion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func resourceProject() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceProjectCreate,
		ReadContext:   resourceProjectRead,
		UpdateContext: resourceProjectUpdate,
		DeleteContext: resourceProjectDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceProjectRead(ctx, d, m)
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			// Notice there is no 'id' field specified because it will be created.
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"archived": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"auto_pay": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"default_aws_region": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"ou_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true, // Not allowed to be changed, forces new item if changed.
			},
			"owner_user_ids": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				AtLeastOneOf: []string{"owner_user_group_ids", "owner_user_ids"},
			},
			"owner_user_group_ids": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Type:         schema.TypeSet,
				Optional:     true,
				Description:  "Must provide at least the owner_user_groups field or the owner_users field.",
				AtLeastOneOf: []string{"owner_user_group_ids", "owner_user_ids"},
			},
			"permission_scheme_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"project_funding": {
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type:     schema.TypeFloat,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"funding_order": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"funding_source_id": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"start_datecode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
						"end_datecode": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true, // Not allowed to be changed, forces new item if changed.
						},
					},
				},
				Type:     schema.TypeSet,
				Optional: true,
			},
			"budget": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"amount": {
							Type: schema.TypeFloat,
							Description: "Total amount for the budget. This is required if data is not specified. " +
								"Budget entries are created between start_datecode and end_datecode (exclusive) with the amount evenly distributed across the months.",
							Optional: true,
						},
						"data": {
							Type: schema.TypeSet,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"datecode": {
										Type:        schema.TypeString,
										Description: "Year and month for the budget data entry (i.e 2023-01).",
										Required:    true,
									},
									"amount": {
										Type:        schema.TypeFloat,
										Description: "Amount of the budget entry in dollars.",
										Required:    true,
									},
									"funding_source_id": {
										Type:        schema.TypeInt,
										Description: "ID of funding source for the budget entry.",
										Optional:    true,
									},
									"priority": {
										Type:        schema.TypeInt,
										Description: "Priority order of the budget entry. This is required if funding_source_id is specified",
										Optional:    true,
									},
								},
							},
							Description: "Total amount for the budget. This is required if data is not specified. " +
								"Budget entries are created between start_datecode and end_datecode (exclusive) with the amount evenly distributed across the months.",
							Optional: true,
						},
						"funding_source_ids": {
							Type: schema.TypeSet,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
							Description: "Funding source IDs to use when data is not specified. " +
								"This value is ignored is data is specified. If specified, the amount is distributed evenly across months and funding sources. " +
								"Funding sources will be processed in order from first to last.",
							Optional: true,
						},
						"start_datecode": {
							Type:        schema.TypeString,
							Description: "Year and month the budget starts.",
							Required:    true,
						},
						"end_datecode": {
							Type:        schema.TypeString,
							Description: "Year and month the budget ends. This is an exclusive date.",
							Required:    true,
						},
					},
				},
			},
			"labels": {
				Type:        schema.TypeMap,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "A map of labels to assign to the project. The labels must already exist in Kion.",
			},
		},
	}
}

func resourceProjectCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	post := hc.ProjectCreate{
		AutoPay:            d.Get("auto_pay").(bool),
		DefaultAwsRegion:   d.Get("default_aws_region").(string),
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		OUID:               d.Get("ou_id").(int),
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_user_ids"),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_group_ids"),
		PermissionSchemeID: d.Get("permission_scheme_id").(int),
	}

	projectCreateURLSuffix := "with-spend-plan"

	// Get financial config settings
	type FinancialConfig struct {
		Data struct {
			BudgetMode bool `json:"budget_mode"`
		} `json:"data"`
	}
	var config FinancialConfig
	err := client.GET("/v1/ct-config/financials-config", &config)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to retrieve financial config",
			Detail:   fmt.Sprintf("Error: %v", err),
		})
		return diags
	}

	// Can't cast directly to []interface{}
	// Must cast each element to map[string]interface{} & assign each value from the POST object.
	if config.Data.BudgetMode {
		projectCreateURLSuffix = "with-budget"

		post.Budget = make([]hc.BudgetCreate, len(d.Get("budget").(*schema.Set).List()))

		for i, genericValue := range d.Get("budget").(*schema.Set).List() {
			// Cast each generic interface{} value to a map of key/value pairs
			budgetMap := genericValue.(map[string]interface{})

			// Unpack struct values & assign them to the POST object
			post.Budget[i] = hc.BudgetCreate{
				Amount:        budgetMap["amount"].(float64),
				StartDatecode: budgetMap["start_datecode"].(string),
				EndDatecode:   budgetMap["end_datecode"].(string),
			}

			if v, ok := budgetMap["funding_source_ids"].(*schema.Set); ok {
				ids := make([]int, 0, v.Len())
				for _, id := range v.List() {
					ids = append(ids, id.(int))
				}
				post.Budget[i].FundingSourceIDs = &ids
			}

			if v, ok := budgetMap["data"].(*schema.Set); ok && v.Len() > 0 {
				post.Budget[i].Data = make([]hc.BudgetDataCreate, v.Len())
				for idx, dataValue := range v.List() {
					dataMap := dataValue.(map[string]interface{})
					post.Budget[i].Data[idx] = hc.BudgetDataCreate{
						Datecode:        dataMap["datecode"].(string),
						Amount:          dataMap["amount"].(float64),
						FundingSourceID: dataMap["funding_source_id"].(int),
						Priority:        dataMap["priority"].(int),
					}
				}
			}
		}
	} else {
		post.ProjectFunding = make([]hc.ProjectFundingCreate, len(d.Get("project_funding").(*schema.Set).List()))

		for i, genericValue := range d.Get("project_funding").(*schema.Set).List() {
			// Cast each generic interface{} value to a map of key/value pairs
			projectFundingMap := genericValue.(map[string]interface{})

			// Unpack struct values & assign them to the POST object
			post.ProjectFunding[i] = hc.ProjectFundingCreate{
				Amount:          projectFundingMap["amount"].(float64),
				FundingOrder:    projectFundingMap["funding_order"].(int),
				FundingSourceID: projectFundingMap["funding_source_id"].(int),
				StartDatecode:   projectFundingMap["start_datecode"].(string),
				EndDatecode:     projectFundingMap["end_datecode"].(string),
			}
		}
	}

	resp, err := client.POST(fmt.Sprintf("/v3/project/%v", projectCreateURLSuffix), post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project",
			Detail:   fmt.Sprintf("Error: %v", err),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project",
			Detail:   "Received item ID of 0",
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	if labels, ok := d.GetOk("labels"); ok && labels != nil {
		ID := d.Id()
		err = hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "project", ID)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Project labels",
				Detail:   fmt.Sprintf("Error: %v", err),
			})
			return diags
		}
	}

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.ProjectResponse)
	err := client.GET(fmt.Sprintf("/v3/project/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project",
			Detail:   fmt.Sprintf("Error: %v", err),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["archived"] = item.Archived
	data["auto_pay"] = item.AutoPay
	data["default_aws_region"] = item.DefaultAwsRegion
	data["description"] = item.Description
	data["name"] = item.Name
	data["ou_id"] = item.OUID

	// Get project budgets
	budgetResp := new(struct {
		Data []struct {
			Config struct {
				ID            int     `json:"id"`
				StartDatecode string  `json:"start_datecode"`
				EndDatecode   string  `json:"end_datecode"`
				Amount        float64 `json:"amount,omitempty"` // This field is not present when using monthly data
			} `json:"config"`
			Data []struct {
				Amount          float64 `json:"amount"`
				Datecode        string  `json:"datecode"`
				FundingSourceID int     `json:"funding_source_id"`
				Priority        int     `json:"priority"`
			} `json:"data"`
		} `json:"data"`
		Status int `json:"status"`
	})
	err = client.GET(fmt.Sprintf("/v3/project/%s/budget", ID), budgetResp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project budgets",
			Detail:   fmt.Sprintf("Error: %v", err),
		})
		return diags
	}

	// Convert budgets to the format expected by the schema
	budgets := make([]map[string]interface{}, 0)
	for _, budget := range budgetResp.Data {
		budgetMap := make(map[string]interface{})
		// Calculate total amount from budget data entries when present
		// The API doesn't return an amount field when using monthly data entries
		var totalAmount float64
		if len(budget.Data) > 0 {
			// Sum up all monthly amounts
			for _, data := range budget.Data {
				totalAmount += data.Amount
			}
		} else {
			// Use the config amount if no monthly data
			totalAmount = budget.Config.Amount
		}
		budgetMap["amount"] = totalAmount
		budgetMap["start_datecode"] = budget.Config.StartDatecode
		budgetMap["end_datecode"] = budget.Config.EndDatecode

		// Extract unique funding source IDs from budget data
		fundingSources := make(map[int]bool)
		for _, data := range budget.Data {
			fundingSources[data.FundingSourceID] = true
		}
		fsIDs := make([]int, 0)
		for fsID := range fundingSources {
			fsIDs = append(fsIDs, fsID)
		}
		budgetMap["funding_source_ids"] = fsIDs

		// Add budget data if present
		if len(budget.Data) > 0 {
			budgetData := make([]map[string]interface{}, len(budget.Data))
			for i, data := range budget.Data {
				budgetData[i] = map[string]interface{}{
					"amount":            data.Amount,
					"datecode":          data.Datecode,
					"funding_source_id": data.FundingSourceID,
					"priority":          data.Priority,
				}
			}
			budgetMap["data"] = budgetData
		}

		budgets = append(budgets, budgetMap)
	}
	data["budget"] = budgets

	for k, v := range data {
		if err := hc.SafeSet(d, k, v, "Unable to read Project"); err != nil {
			diags = append(diags, err...)
			return diags
		}
	}

	// Fetch labels
	labelData, err := hc.ReadResourceLabels(client, "project", ID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project labels",
			Detail:   fmt.Sprintf("Error: %v", err),
		})
		return diags
	}

	// Set labels
	if err := hc.SafeSet(d, "labels", labelData, "Unable to set labels for Project"); err != nil {
		diags = append(diags, err...)
	}

	return diags
}

func resourceProjectUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("archived",
		"auto_pay",
		"default_aws_region",
		"description",
		"name",
		"permission_scheme_id") {
		hasChanged++
		req := hc.ProjectUpdate{
			Archived:           d.Get("archived").(bool),
			AutoPay:            d.Get("auto_pay").(bool),
			DefaultAwsRegion:   d.Get("default_aws_region").(string),
			Description:        d.Get("description").(string),
			Name:               d.Get("name").(string),
			PermissionSchemeID: d.Get("permission_scheme_id").(int),
		}

		err := client.PATCH(fmt.Sprintf("/v3/project/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Handle budget changes
	if d.HasChange("budget") {
		hasChanged++

		// Get current budgets from Kion
		resp := new(struct {
			Data []struct {
				Config struct {
					ID            int    `json:"id"`
					StartDatecode string `json:"start_datecode"`
					EndDatecode   string `json:"end_datecode"`
				} `json:"config"`
			} `json:"data"`
			Status int `json:"status"`
		})
		err := client.GET(fmt.Sprintf("/v3/project/%s/budget", ID), resp)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read Project budgets",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}

		// Track existing budget IDs to determine which ones to delete
		existingBudgetIDs := make(map[int]bool)
		for _, budget := range resp.Data {
			existingBudgetIDs[budget.Config.ID] = true
		}

		// Process new/updated budgets
		type budgetPeriod struct {
			startDate time.Time
			endDate   time.Time
			amount    float64
		}
		budgetPeriods := make([]budgetPeriod, 0)

		for _, genericValue := range d.Get("budget").(*schema.Set).List() {
			budgetMap := genericValue.(map[string]interface{})

			// Validate date format and parse dates
			startDatecode := budgetMap["start_datecode"].(string)
			endDatecode := budgetMap["end_datecode"].(string)

			startDate, err := time.Parse("2006-01", startDatecode)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid start_datecode format",
					Detail:   fmt.Sprintf("start_datecode must be in YYYY-MM format. Got: %s", startDatecode),
				})
				return diags
			}

			endDate, err := time.Parse("2006-01", endDatecode)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid end_datecode format",
					Detail:   fmt.Sprintf("end_datecode must be in YYYY-MM format. Got: %s", endDatecode),
				})
				return diags
			}

			// Check if end date is after start date
			if !endDate.After(startDate) {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Invalid date range",
					Detail:   fmt.Sprintf("end_datecode (%s) must be after start_datecode (%s)", endDatecode, startDatecode),
				})
				return diags
			}

			// Check for overlapping periods
			newPeriod := budgetPeriod{
				startDate: startDate,
				endDate:   endDate,
				amount:    budgetMap["amount"].(float64),
			}

			for _, existing := range budgetPeriods {
				// Check if periods overlap
				if (newPeriod.startDate.Before(existing.endDate) || newPeriod.startDate.Equal(existing.endDate)) &&
					(newPeriod.endDate.After(existing.startDate) || newPeriod.endDate.Equal(existing.startDate)) {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Overlapping budget periods",
						Detail: fmt.Sprintf("Budget period %s to %s overlaps with existing period %s to %s",
							startDatecode, endDatecode,
							existing.startDate.Format("2006-01"), existing.endDate.Format("2006-01")),
					})
					return diags
				}
			}
			budgetPeriods = append(budgetPeriods, newPeriod)

			// Create budget request
			budgetReq := struct {
				Amount           float64               `json:"amount"`
				Data             []hc.BudgetDataCreate `json:"data,omitempty"`
				EndDatecode      string                `json:"end_datecode"`
				FundingSourceIDs *[]int                `json:"funding_source_ids,omitempty"`
				ProjectID        int                   `json:"project_id"`
				StartDatecode    string                `json:"start_datecode"`
			}{
				Amount:        budgetMap["amount"].(float64),
				EndDatecode:   endDatecode,
				StartDatecode: startDatecode,
			}

			// Convert string ID to int for ProjectID
			projectID, err := strconv.Atoi(d.Id())
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to parse Project ID",
					Detail:   fmt.Sprintf("Error: %v\nProject ID: %v", err.Error(), d.Id()),
				})
				return diags
			}
			budgetReq.ProjectID = projectID

			// Handle funding source IDs
			if v, ok := budgetMap["funding_source_ids"].(*schema.Set); ok {
				ids := make([]int, 0, v.Len())
				for _, id := range v.List() {
					ids = append(ids, id.(int))
				}
				budgetReq.FundingSourceIDs = &ids
			}

			// Handle budget data if present
			if v, ok := budgetMap["data"].(*schema.Set); ok && v.Len() > 0 {
				budgetReq.Data = make([]hc.BudgetDataCreate, v.Len())
				for i, dataValue := range v.List() {
					dataMap := dataValue.(map[string]interface{})
					budgetReq.Data[i] = hc.BudgetDataCreate{
						Datecode:        dataMap["datecode"].(string),
						Amount:          dataMap["amount"].(float64),
						FundingSourceID: dataMap["funding_source_id"].(int),
						Priority:        dataMap["priority"].(int),
					}
				}
			}

			// Find if this budget already exists (match by date range)
			var existingBudgetID int
			for _, budget := range resp.Data {
				if budget.Config.StartDatecode == budgetReq.StartDatecode &&
					budget.Config.EndDatecode == budgetReq.EndDatecode {
					existingBudgetID = budget.Config.ID
					delete(existingBudgetIDs, existingBudgetID)
					break
				}
			}

			// Store the request for later processing
			if existingBudgetID != 0 {
				// Update existing budget
				err = client.PUT(fmt.Sprintf("/v3/budget/%d", existingBudgetID), budgetReq)
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Unable to update Project budget",
						Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), existingBudgetID),
					})
					return diags
				}
			}
		}

		// Delete removed budgets first to avoid overlap conflicts
		for budgetID := range existingBudgetIDs {
			err = client.DELETE(fmt.Sprintf("/v3/budget/%d", budgetID), nil)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to delete Project budget",
					Detail:   fmt.Sprintf("Error: %v\nBudget: %v", err.Error(), budgetID),
				})
				return diags
			}
		}

		// Now create new budgets
		for _, genericValue := range d.Get("budget").(*schema.Set).List() {
			budgetMap := genericValue.(map[string]interface{})

			// Skip if this budget already exists (we updated it above)
			startDatecode := budgetMap["start_datecode"].(string)
			endDatecode := budgetMap["end_datecode"].(string)
			exists := false
			for _, budget := range resp.Data {
				if budget.Config.StartDatecode == startDatecode &&
					budget.Config.EndDatecode == endDatecode {
					exists = true
					break
				}
			}
			if exists {
				continue
			}

			// Create budget request for new budget
			budgetReq := struct {
				Amount           float64               `json:"amount"`
				Data             []hc.BudgetDataCreate `json:"data,omitempty"`
				EndDatecode      string                `json:"end_datecode"`
				FundingSourceIDs *[]int                `json:"funding_source_ids,omitempty"`
				ProjectID        int                   `json:"project_id"`
				StartDatecode    string                `json:"start_datecode"`
			}{
				Amount:        budgetMap["amount"].(float64),
				EndDatecode:   endDatecode,
				StartDatecode: startDatecode,
			}

			// Convert string ID to int for ProjectID
			projectID, err := strconv.Atoi(d.Id())
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to parse Project ID",
					Detail:   fmt.Sprintf("Error: %v\nProject ID: %v", err.Error(), d.Id()),
				})
				return diags
			}
			budgetReq.ProjectID = projectID

			// Handle funding source IDs
			if v, ok := budgetMap["funding_source_ids"].(*schema.Set); ok {
				ids := make([]int, 0, v.Len())
				for _, id := range v.List() {
					ids = append(ids, id.(int))
				}
				budgetReq.FundingSourceIDs = &ids
			}

			// Handle budget data if present
			if v, ok := budgetMap["data"].(*schema.Set); ok && v.Len() > 0 {
				budgetReq.Data = make([]hc.BudgetDataCreate, v.Len())
				for i, dataValue := range v.List() {
					dataMap := dataValue.(map[string]interface{})
					budgetReq.Data[i] = hc.BudgetDataCreate{
						Datecode:        dataMap["datecode"].(string),
						Amount:          dataMap["amount"].(float64),
						FundingSourceID: dataMap["funding_source_id"].(int),
						Priority:        dataMap["priority"].(int),
					}
				}
			}

			// Create new budget
			_, err = client.POST("/v3/budget", budgetReq)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to create Project budget",
					Detail:   fmt.Sprintf("Error: %v\nProject: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	// Determine if the owners have changed.
	if d.HasChanges("owner_user_ids",
		"owner_user_group_ids") {
		hasChanged++
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_group_ids")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_user_ids")

		if len(arrAddOwnerUserGroupIds) > 0 ||
			len(arrAddOwnerUserIds) > 0 ||
			len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			_, err := client.POST(fmt.Sprintf("/v1/project/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to change owners on Project",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if d.HasChanges("labels") {
		hasChanged++

		err := hc.PutAppLabelIDs(client, hc.FlattenAssociateLabels(d, "labels"), "project", ID)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update Project labels",
				Detail:   fmt.Sprintf("Error: %v\nProject ID: %v", err.Error(), ID),
			})
			return diags
		}
	}

	if hasChanged > 0 {
		if err := d.Set("last_updated", time.Now().Format(time.RFC850)); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to set last_updated",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	return resourceProjectRead(ctx, d, m)
}

func resourceProjectDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/project/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
