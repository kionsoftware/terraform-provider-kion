package kionclient

// ProjectCloudAccessRoleResponse for: GET /api/v3/project-cloud-access-role/{id}
type ProjectCloudAccessRoleResponse struct {
	Data struct {
		Accounts                  []ObjectWithID `json:"accounts"`
		AwsIamPermissionsBoundary *ObjectWithID  `json:"aws_iam_permissions_boundary"`
		AwsIamPolicies            []ObjectWithID `json:"aws_iam_policies"`
		AzureRoleDefinitions      []ObjectWithID `json:"azure_role_definitions"`
		GCPIamRoles               []ObjectWithID `json:"gcp_iam_roles"`
		ProjectCloudAccessRole    struct {
			ApplyToAllAccounts  bool   `json:"apply_to_all_accounts"`
			AwsIamPath          string `json:"aws_iam_path"`
			AwsIamRoleName      *string `json:"aws_iam_role_name,omitempty"`
			FutureAccounts      bool   `json:"future_accounts"`
			ID                  int    `json:"id"`
			LongTermAccessKeys  bool   `json:"long_term_access_keys"`
			Name                string `json:"name"`
			ProjectID           int    `json:"project_id"`
			ShortTermAccessKeys bool   `json:"short_term_access_keys"`
			WebAccess           bool   `json:"web_access"`
		} `json:"project_cloud_access_role"`
		UserGroups []ObjectWithID `json:"user_groups"`
		Users      []ObjectWithID `json:"users"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectCloudAccessRoleCreate for: POST /api/v3/project-cloud-access-role
type ProjectCloudAccessRoleCreate struct {
	AccountIds                *[]int `json:"account_ids"`
	ApplyToAllAccounts        bool   `json:"apply_to_all_accounts"`
	AwsIamPath                string `json:"aws_iam_path"`
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	AwsIamRoleName            *string `json:"aws_iam_role_name,omitempty"`
	AzureRoleDefinitions      *[]int `json:"azure_role_definitions"`
	FutureAccounts            bool   `json:"future_accounts"`
	GCPIamRoles               *[]int `json:"gcp_iam_roles"`
	LongTermAccessKeys        bool   `json:"long_term_access_keys"`
	Name                      string `json:"name"`
	ProjectID                 int    `json:"project_id"`
	ShortTermAccessKeys       bool   `json:"short_term_access_keys"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
	WebAccess                 bool   `json:"web_access"`
}

// ProjectCloudAccessRoleUpdate for: PATCH /api/v3/project-cloud-access-role/{id}
type ProjectCloudAccessRoleUpdate struct {
	ApplyToAllAccounts  bool   `json:"apply_to_all_accounts"`
	FutureAccounts      bool   `json:"future_accounts"`
	LongTermAccessKeys  bool   `json:"long_term_access_keys"`
	Name                string `json:"name"`
	ShortTermAccessKeys bool   `json:"short_term_access_keys"`
	WebAccess           bool   `json:"web_access"`
}

// ProjectCloudAccessRoleAssociationsAdd for: POST /api/v3/project-cloud-access-role/{id}/association
type ProjectCloudAccessRoleAssociationsAdd struct {
	AccountIds                *[]int `json:"account_ids"`
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	AzureRoleDefinitions      *[]int `json:"azure_role_definitions"`
	GCPIamRoles               *[]int `json:"gcp_iam_roles"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
}

// ProjectCloudAccessRoleAssociationsRemove for: DELETE /api/v3/project-cloud-access-role/{id}/association
type ProjectCloudAccessRoleAssociationsRemove struct {
	AccountIds                *[]int `json:"account_ids"`
	AwsIamPermissionsBoundary *int   `json:"aws_iam_permissions_boundary"`
	AwsIamPolicies            *[]int `json:"aws_iam_policies"`
	AzureRoleDefinitions      *[]int `json:"azure_role_definitions"`
	GCPIamRoles               *[]int `json:"gcp_iam_roles"`
	UserGroupIds              *[]int `json:"user_group_ids"`
	UserIds                   *[]int `json:"user_ids"`
}
