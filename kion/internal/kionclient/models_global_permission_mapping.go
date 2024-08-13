package kionclient

// GlobalPermissionMapping represents a user permission mapping for the entire application.
type GlobalPermissionMapping struct {
	AppRoleID     int   `json:"app_role_id"`
	UserGroupsIDs []int `json:"user_groups_ids"`
	UserIDs       []int `json:"user_ids"`
}

// GlobalPermissionMappingListResponse represents the response from the API for a list of global permission mappings.
type GlobalPermissionMappingListResponse struct {
	Data   []GlobalPermissionMapping `json:"data"`
	Status int                       `json:"status"`
}
