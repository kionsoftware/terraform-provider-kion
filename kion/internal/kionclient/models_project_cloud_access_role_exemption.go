package kionclient

// ProjectCloudAccessRoleExemption defines an exemption of an ou cloud access role for a project
type ProjectCloudAccessRoleExemption struct {
	ID                  int    `json:"id"`
	OUCloudAccessRoleID int    `json:"ou_cloud_access_role_id"`
	ProjectID           int    `json:"project_id"`
	Reason              string `json:"reason"`
}

// ProjectCloudAccessRoleExemptionResponse is the response from GET /v3/project-cloud-access-role-exemption/{id}
type ProjectCloudAccessRoleExemptionResponse struct {
	Data   ProjectCloudAccessRoleExemption `json:"data"`
	Status int                             `json:"status"`
}

// ProjectCloudAccessRoleExemptionListResponse is the response from GET /v3/project-cloud-access-role-exemption
type ProjectCloudAccessRoleExemptionListResponse struct {
	Data struct {
		Items []ProjectCloudAccessRoleExemption `json:"items"`
		Total int                               `json:"total"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectCloudAccessRoleExemptionCreate is the payload for POST /v3/project-cloud-access-role-exemption
type ProjectCloudAccessRoleExemptionCreate struct {
	OUCloudAccessRoleID int    `json:"ou_cloud_access_role_id"`
	ProjectID           int    `json:"project_id"`
	Reason              string `json:"reason"`
}

// ProjectCloudAccessRoleExemptionUpdate is the payload for PATCH /v3/project-cloud-access-role-exemption/{id}
type ProjectCloudAccessRoleExemptionUpdate struct {
	Reason string `json:"reason"`
}

// ProjectCloudAccessRoleExemptionV1 is the structure for v1 API responses
type ProjectCloudAccessRoleExemptionV1 struct {
	ID                  int                 `json:"id"`
	ProjectID           int                 `json:"project_id"`
	OUCloudAccessRoleID int                 `json:"ou_cloud_access_role_id"`
	Reason              string              `json:"reason"`
	OUCloudAccessRole   OUCloudAccessRoleV1 `json:"ou_cloud_access_role"`
}

// OUCloudAccessRoleV1 is the nested structure in v1 API responses
type OUCloudAccessRoleV1 struct {
	ID                       int    `json:"id"`
	Name                     string `json:"name"`
	OUID                     int    `json:"ou_id"`
	CloudAccessRoleTypeID    int    `json:"cloud_access_role_type_id"`
	TrustPolicy              string `json:"trust_policy"`
	AwsIamRoleName           string `json:"aws_iam_role_name"`
	AwsIamPath               string `json:"aws_iam_path"`
	ConsoleAccess            bool   `json:"console_access"`
	ShortTermKeyAccess       bool   `json:"short_term_key_access"`
	LongTermKeyAccess        bool   `json:"long_term_key_access"`
	IsImported               bool   `json:"is_imported"`
	AwsTrustedAccountNumber  string `json:"aws_trusted_account_number"`
	AwsPartition             string `json:"aws_partition"`
	AwsCreateInstanceProfile bool   `json:"aws_create_instance_profile"`
}

// ProjectCloudAccessRoleV1Response is the response from GET /v1/project/{id}/ou-cloud-access-role
type ProjectCloudAccessRoleV1Response struct {
	Status int `json:"status"`
	Data   struct {
		ToKeep            []interface{}                       `json:"to_keep"`
		ToRemove          []interface{}                       `json:"to_remove"`
		ProjectExemptions []ProjectCloudAccessRoleExemptionV1 `json:"project_exemptions"`
		OUExemptions      interface{}                         `json:"ou_exemptions"`
	} `json:"data"`
}
