package kionclient

// FundingSourcePermissionsMapping represents a user permission mapping for a Funding Source.
type FundingSourcePermissionsMapping struct {
	AppRoleID     int   `json:"app_role_id"`
	UserGroupsIDs []int `json:"user_groups_ids"`
	UserIDs       []int `json:"user_ids"`
}

// FundingSourcePermissionsMappingListResponse represents the response from the API for a list of FundingSource permission mappings.
type FundingSourcePermissionsMappingListResponse struct {
	Data   []FundingSourcePermissionsMapping `json:"data"`
	Status int                               `json:"status"`
}
