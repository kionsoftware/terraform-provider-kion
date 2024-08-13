package kionclient

// ProjectPermissionMapping represents a user permission mapping for a project.
type ProjectPermissionMapping struct {
	AppRoleID     int   `json:"app_role_id"`
	UserGroupsIDs []int `json:"user_groups_ids"`
	UserIDs       []int `json:"user_ids"`
}

// ProjectPermissionMappingListResponse represents the response from the API for a list of project permission mappings.
type ProjectPermissionMappingListResponse struct {
	Data   []ProjectPermissionMapping `json:"data"`
	Status int                        `json:"status"`
}
