package kion

import (
	"context"
	"database/sql"
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
			"account_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "AWS account application ID where the AMI is stored.",
			},
			"aws_ami_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Image ID of the AMI from AWS.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description for the AMI in the application.",
			},
			"expiration_alert_number": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The amount of time before the expiration alert is shown.",
			},
			"expiration_alert_unit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The unit of time for the expiration alert (e.g., 'days', 'hours'). This may be null.",
			},
			"expiration_notify": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Will notify the owners that the shared AMI has expired.",
			},
			"expiration_warning_number": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The amount of time before the expiration warning is sent.",
			},
			"expiration_warning_unit": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The unit of time for the expiration warning (e.g., 'days', 'hours'). This may be null.",
			},
			"expires_at": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The expiration date and time of the AMI. This may be null.",
			},
			"last_updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the AMI.",
			},
			"owner_user_group_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of group IDs who will own the AMI. Required if no owner user IDs are listed.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"owner_user_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of user IDs who will own the AMI. Required if no owner group IDs are listed.",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
			"region": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "AWS region where the AMI exists.",
			},
			"sync_deprecation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Will sync the expiration date from the system into the AMI in AWS.",
			},
			"sync_tags": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Will sync the AWS tags from the source AMI into all the accounts where the AMI is shared.",
			},
			"unavailable_in_aws": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Indicates if the AMI is unavailable in AWS.",
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			fields := map[string]interface{}{
				"owner_user_group_ids": d.Get("owner_user_group_ids"),
				"owner_user_ids":       d.Get("owner_user_ids"),
			}
			return hc.AtLeastOneFieldPresent(fields)
		},
	}
}

func resourceAwsAmiCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*hc.Client)

	expiresAtStr := d.Get("expires_at").(string)
	var expiresAt hc.NullTime
	if expiresAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, expiresAtStr)
		if err != nil {
			return hc.HandleError(fmt.Errorf("invalid format for expires_at: %v", err))
		}
		expiresAt = hc.NullTime{
			NullTime: sql.NullTime{
				Time:  parsedTime,
				Valid: true,
			},
		}
	} else {
		expiresAt = hc.NullTime{
			NullTime: sql.NullTime{
				Valid: false,
			},
		}
	}

	expirationAlertUnit := d.Get("expiration_alert_unit").(string)
	var alertUnit hc.NullString
	if expirationAlertUnit != "" {
		alertUnit = hc.NullString{
			NullString: sql.NullString{
				String: expirationAlertUnit,
				Valid:  true,
			},
		}
	} else {
		alertUnit = hc.NullString{
			NullString: sql.NullString{
				Valid: false,
			},
		}
	}

	expirationWarningUnit := d.Get("expiration_warning_unit").(string)
	var warningUnit hc.NullString
	if expirationWarningUnit != "" {
		warningUnit = hc.NullString{
			NullString: sql.NullString{
				String: expirationWarningUnit,
				Valid:  true,
			},
		}
	} else {
		warningUnit = hc.NullString{
			NullString: sql.NullString{
				Valid: false,
			},
		}
	}

	ownerUserGroupIds := d.Get("owner_user_group_ids").([]interface{})
	ownerUserIds := d.Get("owner_user_ids").([]interface{})

	ownerUserGroupIdsInt, err := hc.ConvertInterfaceSliceToIntSlice(ownerUserGroupIds)
	if err != nil {
		return hc.HandleError(fmt.Errorf("error processing owner_user_group_ids: %v", err))
	}

	ownerUserIdsInt, err := hc.ConvertInterfaceSliceToIntSlice(ownerUserIds)
	if err != nil {
		return hc.HandleError(fmt.Errorf("error processing owner_user_ids: %v", err))
	}

	post := hc.AmiCreate{
		AccountID:               d.Get("account_id").(int),
		AwsAmiID:                d.Get("aws_ami_id").(string),
		Description:             d.Get("description").(string),
		Name:                    d.Get("name").(string),
		Region:                  d.Get("region").(string),
		ExpirationAlertNumber:   d.Get("expiration_alert_number").(int),
		ExpirationAlertUnit:     alertUnit,
		ExpirationNotify:        d.Get("expiration_notify").(bool),
		ExpirationWarningNumber: d.Get("expiration_warning_number").(int),
		ExpirationWarningUnit:   warningUnit,
		ExpiresAt:               expiresAt,
		OwnerUserGroupIds:       &ownerUserGroupIdsInt,
		OwnerUserIds:            &ownerUserIdsInt,
		SyncDeprecation:         d.Get("sync_deprecation").(bool),
		SyncTags:                d.Get("sync_tags").(bool),
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
		"name":               item.Name,
		"region":             item.Region,
		"sync_deprecation":   item.SyncDeprecation,
		"sync_tags":          item.SyncTags,
		"unavailable_in_aws": item.UnavailableInAws,
		"owner_user_groups":  hc.InflateObjectWithID(resp.Data.OwnerUserGroups),
		"owner_users":        hc.InflateObjectWithID(resp.Data.OwnerUsers),
	}

	if item.ExpiresAt.Valid {
		data["expires_at"] = item.ExpiresAt.Time.Format(time.RFC3339)
	} else {
		data["expires_at"] = nil
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

	if d.HasChanges("description", "name", "region", "sync_deprecation", "sync_tags", "expires_at", "expiration_alert_number", "expiration_alert_unit", "expiration_notify", "expiration_warning_number", "expiration_warning_unit") {
		hasChanged = true

		// Parse the expires_at field into NullTime
		expiresAtStr := d.Get("expires_at").(string)
		var expiresAt hc.NullTime
		if expiresAtStr != "" {
			parsedTime, err := time.Parse(time.RFC3339, expiresAtStr)
			if err != nil {
				return hc.HandleError(fmt.Errorf("invalid format for expires_at: %v", err))
			}
			expiresAt = hc.NullTime{NullTime: sql.NullTime{Time: parsedTime, Valid: true}}
		} else {
			expiresAt = hc.NullTime{NullTime: sql.NullTime{Valid: false}}
		}

		expirationAlertUnit := d.Get("expiration_alert_unit").(string)
		var alertUnit hc.NullString
		if expirationAlertUnit != "" {
			alertUnit = hc.NullString{
				NullString: sql.NullString{
					String: expirationAlertUnit,
					Valid:  true,
				},
			}
		} else {
			alertUnit = hc.NullString{
				NullString: sql.NullString{
					Valid: false,
				},
			}
		}

		expirationWarningUnit := d.Get("expiration_warning_unit").(string)
		var warningUnit hc.NullString
		if expirationWarningUnit != "" {
			warningUnit = hc.NullString{
				NullString: sql.NullString{
					String: expirationWarningUnit,
					Valid:  true,
				},
			}
		} else {
			warningUnit = hc.NullString{
				NullString: sql.NullString{
					Valid: false,
				},
			}
		}

		req := hc.AmiUpdate{
			Description:             d.Get("description").(string),
			Name:                    d.Get("name").(string),
			Region:                  d.Get("region").(string),
			SyncDeprecation:         d.Get("sync_deprecation").(bool),
			SyncTags:                d.Get("sync_tags").(bool),
			ExpiresAt:               expiresAt,
			ExpirationAlertNumber:   d.Get("expiration_alert_number").(int),
			ExpirationAlertUnit:     alertUnit,
			ExpirationNotify:        d.Get("expiration_notify").(bool),
			ExpirationWarningNumber: d.Get("expiration_warning_number").(int),
			ExpirationWarningUnit:   warningUnit,
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
	client := m.(*hc.Client)
	ID := d.Id()

	err := client.DELETE(fmt.Sprintf("/v3/ami/%s", ID), nil)
	if diags := hc.HandleError(err); diags != nil {
		return diags
	}

	return hc.SafeSet(d, "id", "", "Failed to reset resource ID after deletion")
}
