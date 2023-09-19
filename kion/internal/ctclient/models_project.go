package ctclient

// ProjectListResponse for: GET /api/v3/project
type ProjectListResponse struct {
	Data []struct {
		Archived         bool   `json:"archived"`
		AutoPay          bool   `json:"auto_pay"`
		DefaultAwsRegion string `json:"default_aws_region"`
		Description      string `json:"description"`
		ID               int    `json:"id"`
		Name             string `json:"name"`
		OUID             int    `json:"ou_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectResponse for: GET /api/v3/project/{id}
type ProjectResponse struct {
	Data struct {
		Archived         bool   `json:"archived"`
		AutoPay          bool   `json:"auto_pay"`
		DefaultAwsRegion string `json:"default_aws_region"`
		Description      string `json:"description"`
		ID               int    `json:"id"`
		Name             string `json:"name"`
		OUID             int    `json:"ou_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// ProjectCreate for: POST /api/v3/project
type ProjectCreate struct {
	AutoPay            bool                   `json:"auto_pay"`
	DefaultAwsRegion   string                 `json:"default_aws_region"`
	Description        string                 `json:"description"`
	Name               string                 `json:"name"`
	OUID               int                    `json:"ou_id"`
	OwnerUserGroupIds  *[]int                 `json:"owner_user_group_ids"`
	OwnerUserIds       *[]int                 `json:"owner_user_ids"`
	PermissionSchemeID int                    `json:"permission_scheme_id"`
	ProjectFunding     []ProjectFundingCreate `json:"project_funding,omitempty"`
	Budget             []BudgetCreate         `json:"budget,omitempty"`
}

// ProjectUpdate for: PATCH /api/v3/project/{id}
type ProjectUpdate struct {
	Archived           bool   `json:"archived"`
	AutoPay            bool   `json:"auto_pay"`
	DefaultAwsRegion   string `json:"default_aws_region"`
	Description        string `json:"description"`
	Name               string `json:"name"`
	PermissionSchemeID int    `json:"permission_scheme_id"`
}

// ProjectLabelsResponse for: GET /api/v3/project/{id}/labels
type ProjectLabelsResponse struct {
	Data []struct {
		Color     string `json:"color"`
		ID        int    `json:"id"`
		Key       string `json:"key"`
		ProjectID int    `json:"project_id"`
		Value     string `json:"value"`
	} `json:"data"`
	Status int `json:"status"`
}
