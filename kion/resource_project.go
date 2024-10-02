package kion

import (
	"context"
	"errors"
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
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	}

	// Can't cast directly to []interface{}
	// Must cast each element to map[string]interface{} & assign each value from the map to the POST object.
	if config.Data.BudgetMode {
		projectCreateURLSuffix = "with-budget"

		post.Budget = make([]hc.BudgetCreate, len(d.Get("budget").(*schema.Set).List()))

		for i, genericValue := range d.Get("budget").(*schema.Set).List() {

			// Cast each generic interface{} value to a map of key/value pairs
			budgetMap := genericValue.(map[string]interface{})

			// Unpack struct values & assign them to the POST object
			post.Budget[i] = hc.BudgetCreate{
				Amount:           budgetMap["amount"].(float64),
				FundingSourceIDs: hc.FlattenIntArrayPointer(budgetMap["funding_source_ids"].(*schema.Set).List()),
				StartDatecode:    budgetMap["start_datecode"].(string),
				EndDatecode:      budgetMap["end_datecode"].(string),
			}

			post.Budget[i].Data = make([]hc.BudgetDataCreate, len(budgetMap["data"].(*schema.Set).List()))

			// fill out budget data as needed
			for idx, genericValue2 := range budgetMap["data"].(*schema.Set).List() {

				// Cast each generic interface{} value to a map of key/value pairs
				budgetDataMap := genericValue2.(map[string]interface{})

				post.Budget[i].Data[idx] = hc.BudgetDataCreate{
					Datecode:        budgetDataMap["datecode"].(string),
					Amount:          budgetDataMap["amount"].(float64),
					FundingSourceID: budgetDataMap["funding_source_id"].(int),
					Priority:        budgetDataMap["priority"].(int),
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
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
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
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	resourceProjectRead(ctx, d, m)

	return diags
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
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
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

	for k, v := range data {
		err := d.Set(k, v) // Use assignment instead of short declaration
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set Project",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Fetch labels
	labelData, err := hc.ReadResourceLabels(client, "project", ID)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read Project labels",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// Set labels
	err = d.Set("labels", labelData)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to set labels for Project",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
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
		diags = append(diags, hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")...)
		if len(diags) > 0 {
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
