package kionclient

// IAMPolicyListResponse for: GET /api/v3/iam-policy
type IAMPolicyListResponse struct {
	Data []struct {
		IamPolicy struct {
			AwsIamPath          string `json:"aws_iam_path"`
			AwsManagedPolicy    bool   `json:"aws_managed_policy"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			PathSuffix          string `json:"path_suffix"`
			Policy              string `json:"policy"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"iam_policy"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// IAMPolicyResponse for: GET /api/v3/iam-policy/{id}
type IAMPolicyResponse struct {
	Data struct {
		IamPolicy struct {
			AwsIamPath          string `json:"aws_iam_path"`
			AwsManagedPolicy    bool   `json:"aws_managed_policy"`
			Description         string `json:"description"`
			ID                  int    `json:"id"`
			Name                string `json:"name"`
			PathSuffix          string `json:"path_suffix"`
			Policy              string `json:"policy"`
			SystemManagedPolicy bool   `json:"system_managed_policy"`
		} `json:"iam_policy"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// IAMPolicyCreate for: POST /api/v3/iam-policy
type IAMPolicyCreate struct {
	AwsIamPath        string `json:"aws_iam_path"`
	Description       string `json:"description"`
	Name              string `json:"name"`
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
	Policy            string `json:"policy"`
}

// IAMPolicyUpdate for: PATCH /api/v3/iam-policy/{id}
type IAMPolicyUpdate struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Policy      string `json:"policy"`
}
