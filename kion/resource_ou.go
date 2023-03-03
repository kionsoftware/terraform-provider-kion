package kion

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	hc "github.com/kionsoftware/terraform-provider-kion/kion/internal/ctclient"
)

func resourceOU() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOUCreate,
		ReadContext:   resourceOURead,
		UpdateContext: resourceOUUpdate,
		DeleteContext: resourceOUDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				resourceOURead(ctx, d, m)
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
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
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
				Type:     schema.TypeSet,
				Optional: true,
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
				Type:     schema.TypeSet,
				Optional: true,
			},
			"parent_ou_id": {
				Type:     schema.TypeInt,
				Required: true,
				//ForceNew: true, // Don't let codegen change this.
			},
			"permission_scheme_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
		},
	}
}

func resourceOUCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)

	post := hc.OUCreate{
		Description:        d.Get("description").(string),
		Name:               d.Get("name").(string),
		OwnerUserGroupIds:  hc.FlattenGenericIDPointer(d, "owner_user_groups"),
		OwnerUserIds:       hc.FlattenGenericIDPointer(d, "owner_users"),
		ParentOuID:         d.Get("parent_ou_id").(int),
		PermissionSchemeID: d.Get("permission_scheme_id").(int),
	}

	resp, err := c.POST("/v3/ou", post)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OU",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), post),
		})
		return diags
	} else if resp.RecordID == 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to create OU",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", errors.New("received item ID of 0"), post),
		})
		return diags
	}

	d.SetId(strconv.Itoa(resp.RecordID))

	resourceOURead(ctx, d, m)

	return diags
}

func resourceOURead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	resp := new(hc.OUResponse)
	err := c.GET(fmt.Sprintf("/v3/ou/%s", ID), resp)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to read OU",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}
	item := resp.Data

	data := make(map[string]interface{})
	data["created_at"] = item.OU.CreatedAt
	data["description"] = item.OU.Description
	data["name"] = item.OU.Name
	if hc.InflateObjectWithID(item.OwnerUserGroups) != nil {
		data["owner_user_groups"] = hc.InflateObjectWithID(item.OwnerUserGroups)
	}
	if hc.InflateObjectWithID(item.OwnerUsers) != nil {
		data["owner_users"] = hc.InflateObjectWithID(item.OwnerUsers)
	}
	data["parent_ou_id"] = item.OU.ParentOuID
	data["permission_scheme_id"] = item.OU.PermissionSchemeID

	for k, v := range data {
		if err := d.Set(k, v); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to read and set OU",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	return diags
}

func resourceOUUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	hasChanged := 0

	// Determine if the attributes that are updatable are changed.
	// Leave out fields that are not allowed to be changed like
	// `aws_iam_path` in AWS IAM policies and add `ForceNew: true` to the
	// schema instead.
	if d.HasChanges("description",
		"name",
		"permission_scheme_id") {
		hasChanged++
		req := hc.OUUpdatable{
			Description:        d.Get("description").(string),
			Name:               d.Get("name").(string),
			PermissionSchemeID: d.Get("permission_scheme_id").(int),
		}

		err := c.PATCH(fmt.Sprintf("/v3/ou/%s", ID), req)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to update OU",
				Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
			})
			return diags
		}
	}

	// Allow moving an OU if the parent ID changes and updating permissions.
	// Don't let codegen remove this.
	diags, hasChanged = OUChanges(c, d, diags, hasChanged)
	if len(diags) > 0 {
		return diags
	}

	// Determine if the owners have changed.
	if d.HasChanges("owner_user_groups",
		"owner_users") {
		hasChanged++
		arrAddOwnerUserGroupIds, arrRemoveOwnerUserGroupIds, _, _ := hc.AssociationChanged(d, "owner_user_groups")
		arrAddOwnerUserIds, arrRemoveOwnerUserIds, _, _ := hc.AssociationChanged(d, "owner_users")

		if len(arrAddOwnerUserGroupIds) > 0 ||
			len(arrAddOwnerUserIds) > 0 {
			_, err := c.POST(fmt.Sprintf("/v3/ou/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrAddOwnerUserGroupIds,
				OwnerUserIds:      &arrAddOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to add owners on OU",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}

		if len(arrRemoveOwnerUserGroupIds) > 0 ||
			len(arrRemoveOwnerUserIds) > 0 {
			err := c.DELETE(fmt.Sprintf("/v3/ou/%s/owner", ID), hc.ChangeOwners{
				OwnerUserGroupIds: &arrRemoveOwnerUserGroupIds,
				OwnerUserIds:      &arrRemoveOwnerUserIds,
			})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Unable to remove owners on OU",
					Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
				})
				return diags
			}
		}
	}

	if hasChanged > 0 {
		d.Set("last_updated", time.Now().Format(time.RFC850))
	}

	return resourceOURead(ctx, d, m)
}

func resourceOUDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	c := m.(*hc.Client)
	ID := d.Id()

	err := c.DELETE(fmt.Sprintf("/v2/ou/%s", ID), nil)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Unable to delete OU",
			Detail:   fmt.Sprintf("Error: %v\nItem: %v", err.Error(), ID),
		})
		return diags
	}

	// d.SetId("") is automatically called assuming delete returns no errors, but
	// it is added here for explicitness.
	d.SetId("")

	return diags
}
