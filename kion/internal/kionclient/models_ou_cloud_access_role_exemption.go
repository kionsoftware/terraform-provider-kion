package kionclient

// OUCloudAccessRoleExemption defines an exemption of an ou cloud access role for an ou
type OUCloudAccessRoleExemption struct {
	ID                  int    `json:"id"`
	OUCloudAccessRoleID int    `json:"ou_cloud_access_role_id"`
	OUID                int    `json:"ou_id"`
	Reason              string `json:"reason"`
}

// OUCloudAccessRoleExemptionResponse is the response from GET /v3/ou-cloud-access-role-exemption/{id}
type OUCloudAccessRoleExemptionResponse struct {
	Data   OUCloudAccessRoleExemption `json:"data"`
	Status int                        `json:"status"`
}

// OUCloudAccessRoleExemptionListResponse is the response from GET /v3/ou-cloud-access-role-exemption
type OUCloudAccessRoleExemptionListResponse struct {
	Data struct {
		Items []OUCloudAccessRoleExemption `json:"items"`
		Total int                          `json:"total"`
	} `json:"data"`
	Status int `json:"status"`
}

// OUCloudAccessRoleExemptionCreate is the payload for POST /v3/ou-cloud-access-role-exemption
type OUCloudAccessRoleExemptionCreate struct {
	OUCloudAccessRoleID int    `json:"ou_cloud_access_role_id"`
	OUID                int    `json:"ou_id"`
	Reason              string `json:"reason"`
}

// OUCloudAccessRoleExemptionUpdate is the payload for PATCH /v3/ou-cloud-access-role-exemption/{id}
type OUCloudAccessRoleExemptionUpdate struct {
	Reason string `json:"reason"`
}

// OUCloudAccessRoleExemptionV1 is the structure for v1 API responses
type OUCloudAccessRoleExemptionV1 struct {
	ID                  int                 `json:"id"`
	OUID                int                 `json:"ou_id"`
	OUCloudAccessRoleID int                 `json:"ou_cloud_access_role_id"`
	Reason              string              `json:"reason"`
	OUCloudAccessRole   OUCloudAccessRoleV1 `json:"ou_cloud_access_role"`
}

// OUCloudAccessRoleV1Response is the response from GET /v1/ou/{id}/ou-cloud-access-role
type OUCloudAccessRoleV1Response struct {
	Status int `json:"status"`
	Data   struct {
		ToKeep            []interface{}                  `json:"to_keep"`
		ToRemove          []interface{}                  `json:"to_remove"`
		ProjectExemptions interface{}                    `json:"project_exemptions"`
		OUExemptions      []OUCloudAccessRoleExemptionV1 `json:"ou_exemptions"`
	} `json:"data"`
}
