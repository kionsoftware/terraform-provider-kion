package ctclient

// OUCloudAccessRoleResponse for: GET /api/v3/ou-cloud-access-role/{id}
type OUCloudAccessRoleResponse struct {
	Data struct {
		AwsIamPermissionsBoundary *ObjectWithID  `json:"aws_iam_permissions_boundary"`
		AwsIamPolicies            []ObjectWithID `json:"aws_iam_policies"`
		InheritedAwsIamPolicies   []ObjectWithID `json:"inherited_aws_iam_policies"`
		OUCloudAccessRole         struct {
			AwsIamPath          string `json:"aws_iam_path"`
			AwsIamRoleName      string `json:"aws_iam_role_name"`
			ID                  int    `json:"id"`
			LongTermAccessKeys  bool   `json:"long_term_access_keys"`
			Name                string `json:"name"`
			OUID                int    `json:"ou_id"`
			ShortTermAccessKeys bool   `json:"short_term_access_keys"`
			WebAccess           bool   `json:"web_access"`
		} `json:"ou_cloud_access_role"`
		UserGroups []ObjectWithID `json:"user_groups"`
		Users      []ObjectWithID `json:"users"`
	} `json:"data"`
	Status int `json:"status"`
}

// OUCloudAccessRoleCreate for: POST /api/v3/ou-cloud-access-role
type OUCloudAccessRoleCreate struct {
	AwsIamPath                string `json:"aws_iam_path"`
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	AwsIamRoleName            string `json:"aws_iam_role_name"`
	LongTermAccessKeys        bool   `json:"long_term_access_keys"`
	Name                      string `json:"name"`
	OUID                      int    `json:"ou_id"`
	ShortTermAccessKeys       bool   `json:"short_term_access_keys"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
	WebAccess                 bool   `json:"web_access"`
}

// OUCloudAccessRoleUpdate for: PATCH /api/v3/ou-cloud-access-role/{id}
type OUCloudAccessRoleUpdate struct {
	LongTermAccessKeys  bool   `json:"long_term_access_keys"`
	Name                string `json:"name"`
	ShortTermAccessKeys bool   `json:"short_term_access_keys"`
	WebAccess           bool   `json:"web_access"`
}

// OUCloudAccessRoleAssociationsAdd for: POST /api/v3/ou-cloud-access-role/{id}/association
type OUCloudAccessRoleAssociationsAdd struct {
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
}

// OUCloudAccessRoleAssociationsRemove for: DELETE /api/v3/ou-cloud-access-role/{id}/association
type OUCloudAccessRoleAssociationsRemove struct {
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
}
