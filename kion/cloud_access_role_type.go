package kion

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Cloud access role types recognized by the Kion public API. Every type other
// than User is AWS only and cannot be associated with Kion users, user groups,
// or Azure/GCP role definitions.
const (
	cloudAccessRoleTypeUser        = 1
	cloudAccessRoleTypeCustomTrust = 2
	cloudAccessRoleTypeAccount     = 3
	cloudAccessRoleTypeService     = 4
)

// cloudAccessRoleTypeNames maps a role type to the name used in the Kion UI and
// API documentation, for use in diagnostics.
var cloudAccessRoleTypeNames = map[int]string{
	cloudAccessRoleTypeUser:        "User",
	cloudAccessRoleTypeCustomTrust: "Custom Trust",
	cloudAccessRoleTypeAccount:     "Account",
	cloudAccessRoleTypeService:     "Service",
}

// awsAccountNumberRegex matches a bare 12-digit AWS account number.
var awsAccountNumberRegex = regexp.MustCompile(`^\d{12}$`)

// awsPartitions are the AWS partitions a non-User cloud access role can sync to.
var awsPartitions = []string{"aws", "aws-us-gov", "aws-iso", "aws-iso-b"}

// cloudAccessRoleTypeSchema returns the schema entries shared by the project and
// OU cloud access role resources.
func cloudAccessRoleTypeSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cloud_access_role_type_id": {
			Type:     schema.TypeInt,
			Optional: true,
			Default:  cloudAccessRoleTypeUser,
			// The API exposes no update path for the role type.
			ForceNew:     true,
			ValidateFunc: validation.IntBetween(cloudAccessRoleTypeUser, cloudAccessRoleTypeService),
			Description: "Type of the cloud access role: 1 = User (default), 2 = Custom Trust, " +
				"3 = Account, 4 = Service. Types other than User are AWS only and require Kion 3.15.3, 3.16.5 " +
				"or 3.17.1, depending on the release line. Note that 3.17.0 does not support them. On a version " +
				"without support, Kion creates a User role instead and the provider fails the apply rather than " +
				"letting the mismatch go unnoticed.",
		},
		"aws_iam_role_trust_policy": {
			Type:             schema.TypeString,
			Optional:         true,
			ValidateFunc:     validation.StringIsJSON,
			DiffSuppressFunc: suppressEquivalentJSON,
			Description:      "AWS IAM role trust policy JSON. Required when cloud_access_role_type_id is 2 (Custom Trust).",
		},
		"aws_trusted_account_numbers": {
			Type:     schema.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.StringMatch(awsAccountNumberRegex, "must be a 12-digit AWS account number"),
			},
			Description: "AWS account numbers this role trusts. Required when cloud_access_role_type_id is 3 (Account). " +
				"Kion currently stores exactly one entry.",
		},
		"aws_trusted_services": {
			Type:        schema.TypeSet,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: `AWS service principals this role trusts, such as "lambda.amazonaws.com". Required when cloud_access_role_type_id is 4 (Service).`,
		},
		"aws_partition": {
			Type:         schema.TypeString,
			Optional:     true,
			Computed:     true,
			ValidateFunc: validation.StringInSlice(awsPartitions, false),
			Description: "AWS partition this role is synced to. Applies only to non-User role types and defaults to " +
				`"aws". Callers in GovCloud or ISO partitions must set this explicitly.`,
		},
		"aws_create_instance_profile": {
			Type:        schema.TypeBool,
			Optional:    true,
			Description: "Whether to create an IAM instance profile for this role. Applies only to non-User role types.",
		},
		"aws_session_tags": {
			Type:        schema.TypeMap,
			Optional:    true,
			Elem:        &schema.Schema{Type: schema.TypeString},
			Description: "AWS session tags applied when assuming this role in the AWS console.",
		},
	}
}

// cloudAccessRoleTypeDataSourceSchema returns the same fields as
// cloudAccessRoleTypeSchema, as read-only data source attributes.
func cloudAccessRoleTypeDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"cloud_access_role_type_id": {
			Type:     schema.TypeInt,
			Computed: true,
		},
		"aws_iam_role_trust_policy": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"aws_trusted_account_numbers": {
			Type:     schema.TypeList,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"aws_trusted_services": {
			Type:     schema.TypeSet,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
		"aws_partition": {
			Type:     schema.TypeString,
			Computed: true,
		},
		"aws_create_instance_profile": {
			Type:     schema.TypeBool,
			Computed: true,
		},
		"aws_session_tags": {
			Type:     schema.TypeMap,
			Computed: true,
			Elem:     &schema.Schema{Type: schema.TypeString},
		},
	}
}

// mergeSchemas copies every entry of src into dst, panicking on a duplicate key
// so a collision surfaces at provider start rather than at apply time.
func mergeSchemas(dst, src map[string]*schema.Schema) map[string]*schema.Schema {
	for k, v := range src {
		if _, exists := dst[k]; exists {
			panic(fmt.Sprintf("duplicate schema key %q", k))
		}
		dst[k] = v
	}
	return dst
}

// suppressEquivalentJSON reports whether two JSON documents are semantically
// equal, so that re-formatting by the API does not show up as a diff.
func suppressEquivalentJSON(_, old, new string, _ *schema.ResourceData) bool {
	if old == new {
		return true
	}
	if old == "" || new == "" {
		return false
	}

	var oldDoc, newDoc interface{}
	if err := json.Unmarshal([]byte(old), &oldDoc); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(new), &newDoc); err != nil {
		return false
	}

	return reflect.DeepEqual(oldDoc, newDoc)
}

// cloudAccessRoleTypeConfig is the subset of a cloud access role configuration
// that the role type rules depend on.
type cloudAccessRoleTypeConfig struct {
	roleType        int
	hasTrustPolicy  bool
	trustedAccounts int
	trustedServices int
	users           int
	userGroups      int
	azureRoles      int
	gcpRoles        int
}

// validateCloudAccessRoleTypeDiff enforces the per-type field rules the Kion API
// applies, so that a misconfiguration fails at plan time with a specific message
// instead of as a generic 400 at apply time.
func validateCloudAccessRoleTypeDiff(_ context.Context, d *schema.ResourceDiff, _ interface{}) error {
	trustPolicy, hasTrustPolicy := d.GetOk("aws_iam_role_trust_policy")

	cfg := cloudAccessRoleTypeConfig{
		roleType:       d.Get("cloud_access_role_type_id").(int),
		hasTrustPolicy: hasTrustPolicy && trustPolicy.(string) != "",
	}

	if accounts, ok := d.Get("aws_trusted_account_numbers").([]interface{}); ok {
		cfg.trustedAccounts = len(accounts)
	}

	for field, count := range map[string]*int{
		"aws_trusted_services":   &cfg.trustedServices,
		"users":                  &cfg.users,
		"user_groups":            &cfg.userGroups,
		"azure_role_definitions": &cfg.azureRoles,
		"gcp_iam_roles":          &cfg.gcpRoles,
	} {
		if set, ok := d.Get(field).(*schema.Set); ok {
			*count = set.Len()
		}
	}

	return validateCloudAccessRoleTypeConfig(cfg)
}

// validateCloudAccessRoleTypeConfig holds the role type rules themselves, split
// out from schema plumbing so they can be exercised directly.
func validateCloudAccessRoleTypeConfig(cfg cloudAccessRoleTypeConfig) error {
	switch cfg.roleType {
	case cloudAccessRoleTypeUser:
		if cfg.hasTrustPolicy || cfg.trustedAccounts > 0 || cfg.trustedServices > 0 {
			return fmt.Errorf("aws_iam_role_trust_policy, aws_trusted_account_numbers, and aws_trusted_services " +
				"cannot be set on a User cloud access role (cloud_access_role_type_id = 1)")
		}
		// The remaining rules constrain the AWS-only types.
		return nil

	case cloudAccessRoleTypeCustomTrust:
		if !cfg.hasTrustPolicy {
			return fmt.Errorf("aws_iam_role_trust_policy is required when cloud_access_role_type_id is %d (Custom Trust)",
				cloudAccessRoleTypeCustomTrust)
		}
		if cfg.trustedAccounts > 0 || cfg.trustedServices > 0 {
			return fmt.Errorf("aws_trusted_account_numbers and aws_trusted_services cannot be set on a Custom Trust cloud access role")
		}

	case cloudAccessRoleTypeAccount:
		if cfg.trustedAccounts == 0 {
			return fmt.Errorf("aws_trusted_account_numbers is required when cloud_access_role_type_id is %d (Account)",
				cloudAccessRoleTypeAccount)
		}
		if cfg.hasTrustPolicy || cfg.trustedServices > 0 {
			return fmt.Errorf("aws_iam_role_trust_policy and aws_trusted_services cannot be set on an Account cloud access role")
		}

	case cloudAccessRoleTypeService:
		if cfg.trustedServices == 0 {
			return fmt.Errorf("aws_trusted_services is required when cloud_access_role_type_id is %d (Service)",
				cloudAccessRoleTypeService)
		}
		if cfg.hasTrustPolicy || cfg.trustedAccounts > 0 {
			return fmt.Errorf("aws_iam_role_trust_policy and aws_trusted_account_numbers cannot be set on a Service cloud access role")
		}
	}

	typeName := cloudAccessRoleTypeNames[cfg.roleType]

	if cfg.users > 0 || cfg.userGroups > 0 {
		return fmt.Errorf("users and user_groups cannot be associated with a %s cloud access role; "+
			"only User roles accept Kion users and user groups", typeName)
	}

	if cfg.azureRoles > 0 || cfg.gcpRoles > 0 {
		return fmt.Errorf("azure_role_definitions and gcp_iam_roles cannot be associated with a %s cloud access role, which is AWS only",
			typeName)
	}

	return nil
}

// verifyCloudAccessRoleType confirms Kion honored a non-User role type request.
//
// Kion decodes public API request bodies with json.NewDecoder and does not call
// DisallowUnknownFields, so an installation older than 3.16.5 accepts
// cloud_access_role_type_id, discards it, and creates a User role. Without this
// check that failure is silent and surfaces only as a permanent diff.
func verifyCloudAccessRoleType(resourceName string, requested, actual int) diag.Diagnostics {
	if requested <= cloudAccessRoleTypeUser || requested == actual {
		return nil
	}

	requestedName := cloudAccessRoleTypeNames[requested]
	detail := fmt.Sprintf(
		"Requested cloud_access_role_type_id %d (%s) but Kion returned %d. Cloud access role types reached the "+
			"public API in 3.15.3, 3.16.5 and 3.17.1, depending on the release line; earlier versions, 3.17.0 "+
			"included, ignore the field and create a User role. The role does exist in Kion as a User role, and "+
			"Terraform has marked it tainted, so the next apply will destroy and recreate it. Either upgrade Kion "+
			"or set cloud_access_role_type_id to %d.",
		requested, requestedName, actual, cloudAccessRoleTypeUser)

	return diag.Diagnostics{{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("Kion did not apply the requested %s cloud access role type", resourceName),
		Detail:   detail,
	}}
}

// normalizeCloudAccessRoleTypeID maps the zero value returned by Kion versions
// without role type support onto the User type, so that state matches the
// schema default instead of producing a diff.
func normalizeCloudAccessRoleTypeID(id int) int {
	if id == 0 {
		return cloudAccessRoleTypeUser
	}
	return id
}
