package kionclient

// GCPRoleListResponseWithOwners for: GET /api/v3/gcp-iam-role
type GCPRoleListResponseWithOwners struct {
	Data []struct {
		GcpRole struct {
			ID                  int      `json:"id"`
			GCPID               string   `json:"gcp_id"`
			GCPRoleLaunchStage  int      `json:"gcp_role_launch_stage"`
			Name                string   `json:"name"`
			Description         string   `json:"description"`
			RolePermissions     []string `json:"role_permissions"`
			GCPManagedPolicy    bool     `json:"gcp_managed_policy"`
			SystemManagedPolicy bool     `json:"system_managed_policy"`
		} `json:"gcp_role"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// GCPRoleResponseWithOwners for: GET /api/v3/gcp-iam-role/{id}
type GCPRoleResponseWithOwners struct {
	Data struct {
		GcpRole struct {
			ID                  int      `json:"id"`
			GCPID               string   `json:"gcp_id"`
			GCPRoleLaunchStage  int      `json:"gcp_role_launch_stage"`
			Name                string   `json:"name"`
			Description         string   `json:"description"`
			RolePermissions     []string `json:"role_permissions"`
			GCPManagedPolicy    bool     `json:"gcp_managed_policy"`
			SystemManagedPolicy bool     `json:"system_managed_policy"`
		} `json:"gcp_role"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// GCPRoleCreate for: POST /api/v3/gcp-iam-role
type GCPRoleCreate struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	GCPRoleLaunchStage int      `json:"gcp_role_launch_stage"`
	OwnerUserIDs       *[]int   `json:"owner_user_ids"`
	OwnerUGroupIDs     *[]int   `json:"owner_user_group_ids"`
	RolePermissions    []string `json:"role_permissions"`
}

// GCPRoleUpdate for: PATCH /api/v3/gcp-iam-role/{id}
type GCPRoleUpdate struct {
	Name               string   `json:"name"`
	Description        string   `json:"description"`
	RolePermissions    []string `json:"role_permissions"`
	GCPRoleLaunchStage int      `json:"gcp_role_launch_stage"`
}
