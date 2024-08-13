package kionclient

// OUPermissionMapping represents a user permission mapping for an organizational unit (OU).
type OUPermissionMapping struct {
	AppRoleID     int   `json:"app_role_id"`
	UserGroupsIDs []int `json:"user_groups_ids"`
	UserIDs       []int `json:"user_ids"`
}

// OUPermissionMappingListResponse represents the response from the API for a list of OU permission mappings.
type OUPermissionMappingListResponse struct {
	Data   []OUPermissionMapping `json:"data"`
	Status int                   `json:"status"`
}
