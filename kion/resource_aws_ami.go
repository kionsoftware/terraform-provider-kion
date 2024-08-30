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

func resourceAwsAmi() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAwsAmiCreate,
		ReadContext:   resourceAwsAmiRead,
		UpdateContext: resourceAwsAmiUpdate,
		DeleteContext: resourceAwsAmiDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"last_updated": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"account_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"aws_ami_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"expires_at": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"region": {
				Type:     schema.TypeString,
				Required: true,
			},
			"sync_deprecation": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"sync_tags": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"unavailable_in_aws": {
				Type:     schema.TypeBool,
				Optional: true,
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
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
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
				Type:         schema.TypeSet,
				Optional:     true,
				AtLeastOneOf: []string{"owner_user_groups", "owner_users"},
			},
		},
	}
}

func resourceAwsAmiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	post := hc.AmiCreate{
		AccountID:         d.Get("account_id").(int),
		AwsAmiID:          d.Get("aws_ami_id").(string),
		Description:       d.Get("description").(string),
		Name:              d.Get("name").(string),
		Region:            d.Get("region").(string),
		ExpiresAt:         d.Get("expires_at").(string),
		SyncDeprecation:   d.Get("sync_deprecation").(bool),
		SyncTags:          d.Get("sync_tags").(bool),
		OwnerUserGroupIds: hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:      hc.FlattenGenericIDPointer(d, "owner_users"),
	}

	resp, err := client.POST("/v3/ami", post)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to create AWS AMI: %v", err))
	} else if resp.RecordID == 0 {
		return hc.HandleError(fmt.Errorf("unable to create AWS AMI: received item ID of 0"))
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	return resourceAwsAmiRead(ctx, d, m)
}

func resourceAwsAmiRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.AmiResponse)
	err := client.GET(fmt.Sprintf("/v3/ami/%s", ID), resp)
	if err != nil {
		return hc.HandleError(fmt.Errorf("unable to read AWS AMI: %v", err))
	}
	item := resp.Data.Ami

	data := map[string]interface{}{
		"account_id":         item.AccountID,
		"aws_ami_id":         item.AwsAmiID,
		"description":        item.Description,
		"expires_at":         item.ExpiresAt.Format(time.RFC3339),
		"name":               item.Name,
		"region":             item.Region,
		"sync_deprecation":   item.SyncDeprecation,
		"sync_tags":          item.SyncTags,
		"unavailable_in_aws": item.UnavailableInAws,
		"owner_user_groups":  hc.InflateObjectWithID(resp.Data.OwnerUserGroups),
		"owner_users":        hc.InflateObjectWithID(resp.Data.OwnerUsers),
	}

	for k, v := range data {
		diags = append(diags, hc.SafeSet(d, k, v, "Unable to read and set AWS AMI")...)
	}

	return diags
}

func resourceAwsAmiUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	hasChanged := false

	if d.HasChanges("description", "name", "region", "sync_deprecation", "sync_tags", "expires_at") {
		hasChanged = true
		req := hc.AmiUpdate{
			Description:     d.Get("description").(string),
			Name:            d.Get("name").(string),
			Region:          d.Get("region").(string),
			SyncDeprecation: d.Get("sync_deprecation").(bool),
			SyncTags:        d.Get("sync_tags").(bool),
			ExpiresAt:       d.Get("expires_at").(string),
		}

		err := client.PATCH(fmt.Sprintf("/v3/ami/%s", ID), req)
		if err != nil {
			return hc.HandleError(fmt.Errorf("unable to update AWS AMI: %v", err))
		}
	}

	if d.HasChanges("owner_user_groups", "owner_users") {
		hasChanged = true
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_groups")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_users")

		if len(arrAddOwnerUserGroupIds) > 0 || len(arrAddOwnerUserIds) > 0 {
			_, err := client.POST(fmt.Sprintf("/v3/ami/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				return hc.HandleError(fmt.Errorf("unable to add owners on AWS AMI: %v", err))
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 || len(arrRemoveOwnerUserIds) > 0 {
			err := client.DELETE(fmt.Sprintf("/v3/ami/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				return hc.HandleError(fmt.Errorf("unable to remove owners on AWS AMI: %v", err))
			}
		}
	}

	if hasChanged {
		diags = hc.SafeSet(d, "last_updated", time.Now().Format(time.RFC850), "Failed to set last_updated")
		if len(diags) > 0 {
			return diags
		}
	}

	return resourceAwsAmiRead(ctx, d, m)
}

func resourceAwsAmiDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/ami/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete AWS AMI",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	d.SetId("")

	return diags
}
