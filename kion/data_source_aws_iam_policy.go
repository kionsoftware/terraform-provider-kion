package kion

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/kionclient"
)

func dataSourceAwsIamPolicy() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceAwsIamPolicyRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Description: "The field name whose values you wish to filter by.",
							Type:        schema.TypeString,
							Required:    true,
						},
						"values": {
							Description: "The values of the field name you specified.",
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"regex": {
							Description: "Dictates if the values provided should be treated as regular expressions.",
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
						},
					},
				},
			},
			"list": {
				Description: "This is where Kion makes the discovered data available as a list of resources.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"aws_iam_path": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"aws_managed_policy": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"description": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"owner_user_groups": {
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							Type:     schema.TypeList,
							Computed: true,
						},
						"owner_users": {
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeInt,
										Optional: true,
									},
								},
							},
							Type:     schema.TypeList,
							Computed: true,
						},
						"path_suffix": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"policy": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"system_managed_policy": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"query": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Query string for IAM policy name matching",
			},
			"policy_type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Policy type filter. Valid values are 'user', 'aws', or 'system'",
				ValidateFunc: validation.StringInSlice([]string{
					"user",
					"aws",
					"system",
				}, false),
			},
			"page": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Page number of results",
			},
			"page_size": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of results per page",
			},
		},
	}
}

func dataSourceAwsIamPolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)

	respV4 := new(hc.IAMPolicyV4ListResponse)
	params := make(map[string]string)
	if v, ok := d.GetOk("query"); ok {
		params["query"] = v.(string)
	}
	if v, ok := d.GetOk("policy_type"); ok {
		params["policy-type"] = v.(string)
	}
	if v, ok := d.GetOk("page"); ok {
		params["page"] = strconv.Itoa(v.(int))
		if v2, ok := d.GetOk("page_size"); ok {
			params["count"] = strconv.Itoa(v2.(int))
		} else {
			params["count"] = "100" // default page size
		}
	}

	err := client.GETWithParams("/v4/iam-policy", params, respV4)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	return processV4Response(d, respV4)
}

func processV4Response(d *schema.ResourceData, resp *hc.IAMPolicyV4ListResponse) diag.Diagnostics {
	var diags diag.Diagnostics
	f := hc.NewFilterable(d)

	arr := make([]map[string]interface{}, 0)
	for _, item := range resp.Data.Items {
		data := make(map[string]interface{})
		data["aws_iam_path"] = item.IamPolicy.AwsIamPath
		data["aws_managed_policy"] = item.IamPolicy.AwsManagedPolicy
		data["description"] = item.IamPolicy.Description
		data["id"] = item.IamPolicy.ID
		data["name"] = item.IamPolicy.Name
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
		data["path_suffix"] = item.IamPolicy.PathSuffix
		data["policy"] = item.IamPolicy.Policy
		data["system_managed_policy"] = item.IamPolicy.SystemManagedPolicy

		match, err := f.Match(data)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to filter AwsIamPolicy",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "filter"),
			})
			return diags
		} else if !match {
			continue
		}

		arr = append(arr, data)
	}

	if err := d.Set("list", arr); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read AwsIamPolicy",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), "all"),
		})
		return diags
	}

	// Always run.
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
