package ctclient

// ServiceControlPolicyListResponse for: GET /api/v3/service-control-policy
type ServiceControlPolicyListResponse struct {
	Data []struct {
		OwnerUserGroups      []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers           []ObjectWithID `json:"owner_users"`
		ServiceControlPolicy struct {
			AwsManagedPolicy    bool   `json:"aws_managed_policy"`
			CreatedByUserID     int    `json:"created_by_user_id"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			Policy              string `json:"policy"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"service_control_policy"`
	} `json:"data"`
	Status int `json:"status"`
}

// ServiceControlPolicyResponse for: GET /api/v3/service-control-policy/{id}
type ServiceControlPolicyResponse struct {
	Data struct {
		OwnerUserGroups      []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers           []ObjectWithID `json:"owner_users"`
		ServiceControlPolicy struct {
			AwsManagedPolicy    bool   `json:"aws_managed_policy"`
			CreatedByUserID     int    `json:"created_by_user_id"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			Policy              string `json:"policy"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"service_control_policy"`
	} `json:"data"`
	Status int `json:"status"`
}

// ServiceControlPolicyCreate for: POST /api/v3/service-control-policy
type ServiceControlPolicyCreate struct {
	Description       string `json:"description"`
	Name              string `json:"name"`
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
	Policy            string `json:"policy"`
}

// ServiceControlPolicyUpdate for: PATCH /api/v3/service-control-policy/{id}
type ServiceControlPolicyUpdate struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Policy      string `json:"policy"`
}
