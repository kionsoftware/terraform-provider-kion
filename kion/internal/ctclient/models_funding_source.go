package ctclient

// FundingSourceListResponse for: GET /api/v3/funding-source
type FundingSourceListResponse struct {
	Data []struct {
		ID            int    `json:"id"`
		Amount        int    `json:"amount"`
		Description   string `json:"description"`
		EndDatecode   string `json:"end_datecode"`
		Name          string `json:"name"`
		OUID          int    `json:"ou_id"`
		StartDatecode string `json:"start_datecode"`
	} `json:"data"`
	Status int `json:"status"`
}

// FundingSourceResponse for: GET /api/v3/funding-source/{id}
type FundingSourceResponse struct {
	Data struct {
		ID                 int    `json:"id"`
		Amount             int    `json:"amount"`
		Description        string `json:"description"`
		EndDatecode        string `json:"end_datecode"`
		Name               string `json:"name"`
		OUID               int    `json:"ou_id"`
		StartDatecode      string `json:"start_datecode"`
		PermissionSchemeID int    `json:"permission_scheme_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// FundingSourceCreate for: POST /api/v3/funding-source
type FundingSourceCreate struct {
	ID                 int    `json:"id"`
	Amount             int    `json:"amount"`
	Description        string `json:"description"`
	EndDatecode        string `json:"end_datecode"`
	Name               string `json:"name"`
	StartDatecode      string `json:"start_datecode"`
	PermissionSchemeID int    `json:"permission_scheme_id"`
	OUID               int    `json:"ou_id"`
	OwnerUserGroupIds  *[]int `json:"owner_user_group_ids"`
	OwnerUserIds       *[]int `json:"owner_user_ids"`
}

// FundingSourceUpdate for: PATCH /api/v3/funding-source/{id}
type FundingSourceUpdate struct {
	ID            int    `json:"id"`
	Amount        int    `json:"amount"`
	Description   string `json:"description"`
	EndDatecode   string `json:"end_datecode"`
	Name          string `json:"name"`
	OUID          int    `json:"ou_id"`
	StartDatecode string `json:"start_datecode"`
}

// FundingSourcePermissionMapping
type FundingSourcePermissionMapping struct {
	AppRoleID    int    `json:"app_role_id"`
	UserGroupIds *[]int `json:"user_groups_ids"`
	UserIds      *[]int `json:"user_ids"`
}

type FSUserMappingListResponse struct {
	Data []struct {
		AppRoleId    int    `json:"app_role_id"`
		UserGroupIds *[]int `json:"user_groups_ids"`
		UserIds      *[]int `json:"user_ids"`
	}
}
