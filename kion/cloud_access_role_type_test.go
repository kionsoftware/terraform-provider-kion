package kion

import (
	"strings"
	"testing"
)

func TestNormalizeCloudAccessRoleTypeID(t *testing.T) {
	tests := map[string]struct {
		id   int
		want int
	}{
		"zero from a Kion older than 3.16.5 becomes User": {id: 0, want: cloudAccessRoleTypeUser},
		"User is unchanged":         {id: cloudAccessRoleTypeUser, want: cloudAccessRoleTypeUser},
		"Custom Trust is unchanged": {id: cloudAccessRoleTypeCustomTrust, want: cloudAccessRoleTypeCustomTrust},
		"Service is unchanged":      {id: cloudAccessRoleTypeService, want: cloudAccessRoleTypeService},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := normalizeCloudAccessRoleTypeID(tc.id); got != tc.want {
				t.Errorf("normalizeCloudAccessRoleTypeID(%d) = %d, want %d", tc.id, got, tc.want)
			}
		})
	}
}

func TestVerifyCloudAccessRoleType(t *testing.T) {
	tests := map[string]struct {
		requested int
		actual    int
		wantError bool
	}{
		"User request is never verified": {
			requested: cloudAccessRoleTypeUser, actual: cloudAccessRoleTypeUser, wantError: false,
		},
		"User request against a Kion without type support is tolerated": {
			// Pre-3.16.5 reads back as 0, normalized to User before it gets
			// here. Guard against a raw 0 regardless.
			requested: cloudAccessRoleTypeUser, actual: 0, wantError: false,
		},
		"honored Custom Trust request passes": {
			requested: cloudAccessRoleTypeCustomTrust, actual: cloudAccessRoleTypeCustomTrust, wantError: false,
		},
		"silently downgraded Custom Trust request fails": {
			requested: cloudAccessRoleTypeCustomTrust, actual: cloudAccessRoleTypeUser, wantError: true,
		},
		"silently downgraded Service request fails": {
			requested: cloudAccessRoleTypeService, actual: cloudAccessRoleTypeUser, wantError: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			diags := verifyCloudAccessRoleType("project", tc.requested, tc.actual)

			if got := diags.HasError(); got != tc.wantError {
				t.Fatalf("verifyCloudAccessRoleType(%d, %d) error = %v, want %v",
					tc.requested, tc.actual, got, tc.wantError)
			}

			if tc.wantError && !strings.Contains(diags[0].Detail, "3.16.5") {
				t.Errorf("expected the diagnostic to name the minimum Kion version, got: %s", diags[0].Detail)
			}
		})
	}
}

func TestValidateCloudAccessRoleTypeConfig(t *testing.T) {
	tests := map[string]struct {
		cfg       cloudAccessRoleTypeConfig
		wantError string
	}{
		"plain User role": {
			cfg: cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeUser, users: 2, azureRoles: 1},
		},
		"User role with a trust policy": {
			cfg:       cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeUser, hasTrustPolicy: true},
			wantError: "cannot be set on a User cloud access role",
		},
		"User role with trusted services": {
			cfg:       cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeUser, trustedServices: 1},
			wantError: "cannot be set on a User cloud access role",
		},
		"Custom Trust role with a policy": {
			cfg: cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeCustomTrust, hasTrustPolicy: true},
		},
		"Custom Trust role without a policy": {
			cfg:       cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeCustomTrust},
			wantError: "aws_iam_role_trust_policy is required",
		},
		"Custom Trust role with trusted accounts": {
			cfg: cloudAccessRoleTypeConfig{
				roleType: cloudAccessRoleTypeCustomTrust, hasTrustPolicy: true, trustedAccounts: 1,
			},
			wantError: "cannot be set on a Custom Trust cloud access role",
		},
		"Account role with a trusted account": {
			cfg: cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeAccount, trustedAccounts: 1},
		},
		"Account role without a trusted account": {
			cfg:       cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeAccount},
			wantError: "aws_trusted_account_numbers is required",
		},
		"Service role with a trusted service": {
			cfg: cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeService, trustedServices: 1},
		},
		"Service role without a trusted service": {
			cfg:       cloudAccessRoleTypeConfig{roleType: cloudAccessRoleTypeService},
			wantError: "aws_trusted_services is required",
		},
		"non-User role with users attached": {
			cfg: cloudAccessRoleTypeConfig{
				roleType: cloudAccessRoleTypeAccount, trustedAccounts: 1, users: 1,
			},
			wantError: "only User roles accept Kion users and user groups",
		},
		"non-User role with user groups attached": {
			cfg: cloudAccessRoleTypeConfig{
				roleType: cloudAccessRoleTypeAccount, trustedAccounts: 1, userGroups: 1,
			},
			wantError: "only User roles accept Kion users and user groups",
		},
		"non-User role with Azure roles attached": {
			cfg: cloudAccessRoleTypeConfig{
				roleType: cloudAccessRoleTypeService, trustedServices: 1, azureRoles: 1,
			},
			wantError: "which is AWS only",
		},
		"non-User role with GCP roles attached": {
			cfg: cloudAccessRoleTypeConfig{
				roleType: cloudAccessRoleTypeService, trustedServices: 1, gcpRoles: 1,
			},
			wantError: "which is AWS only",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := validateCloudAccessRoleTypeConfig(tc.cfg)

			if tc.wantError == "" {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("expected an error containing %q, got none", tc.wantError)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("expected an error containing %q, got: %v", tc.wantError, err)
			}
		})
	}
}

func TestSuppressEquivalentJSON(t *testing.T) {
	const policy = `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"sts:AssumeRole"}]}`

	tests := map[string]struct {
		old, new string
		want     bool
	}{
		"identical":                     {old: policy, new: policy, want: true},
		"reformatted":                   {old: policy, new: "  {\n\"Version\" : \"2012-10-17\",\n\"Statement\":[{\"Action\":\"sts:AssumeRole\",\"Effect\":\"Allow\"}]}", want: true},
		"different":                     {old: policy, new: `{"Version":"2012-10-17","Statement":[]}`, want: false},
		"old empty":                     {old: "", new: policy, want: false},
		"new empty":                     {old: policy, new: "", want: false},
		"both empty":                    {old: "", new: "", want: true},
		"unparseable is not suppressed": {old: policy, new: "{not json", want: false},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if got := suppressEquivalentJSON("aws_iam_role_trust_policy", tc.old, tc.new, nil); got != tc.want {
				t.Errorf("suppressEquivalentJSON(%q, %q) = %v, want %v", tc.old, tc.new, got, tc.want)
			}
		})
	}
}

func TestMergeSchemasRejectsDuplicates(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("expected mergeSchemas to panic on a duplicate key")
		}
	}()

	base := cloudAccessRoleTypeSchema()
	mergeSchemas(base, cloudAccessRoleTypeSchema())
}
