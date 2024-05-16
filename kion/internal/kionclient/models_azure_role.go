package kionclient

// AzureRoleListResponse for: GET /api/v3/azure-role
type AzureRoleListResponse struct {
	Data []struct {
		AzureRole struct {
			AzureManagedPolicy  bool   `json:"azure_managed_policy"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			RolePermissions     string `json:"role_permissions"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"azure_role"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzureRoleResponse for: GET /api/v3/azure-role/{id}
type AzureRoleResponse struct {
	Data struct {
		AzureRole struct {
			AzureManagedPolicy  bool   `json:"azure_managed_policy"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			RolePermissions     string `json:"role_permissions"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"azure_role"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzureRoleCreate for: POST /api/v3/azure-role
type AzureRoleCreate struct {
	Description       string `json:"description"`
	Name              string `json:"name"`
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
	RolePermissions   string `json:"role_permissions"`
}

// AzureRoleUpdate for: PATCH /api/v3/azure-role/{id}
type AzureRoleUpdate struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	RolePermissions string `json:"role_permissions"`
}
