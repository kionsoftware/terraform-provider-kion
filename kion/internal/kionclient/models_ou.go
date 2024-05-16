package kionclient

// OUListResponse for: GET /api/v3/ou
type OUListResponse struct {
	Data []struct {
		CreatedAt          string `json:"created_at"`
		Description        string `json:"description"`
		ID                 int    `json:"id"`
		Name               string `json:"name"`
		ParentOuID         int    `json:"parent_ou_id"`
		PermissionSchemeID int    `json:"permission_scheme_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// OUResponse for: GET /api/v3/ou/{id}
type OUResponse struct {
	Data struct {
		OU struct {
			CreatedAt          string `json:"created_at"`
			Description        string `json:"description"`
			ID                 int    `json:"id"`
			Name               string `json:"name"`
			ParentOuID         int    `json:"parent_ou_id"`
			PermissionSchemeID int    `json:"permission_scheme_id"`
		} `json:"ou"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// OUCreate for: POST /api/v3/ou
type OUCreate struct {
	Description        string `json:"description"`
	Name               string `json:"name"`
	OwnerUserGroupIds  *[]int `json:"owner_user_group_ids"`
	OwnerUserIds       *[]int `json:"owner_user_ids"`
	ParentOuID         int    `json:"parent_ou_id"`
	PermissionSchemeID int    `json:"permission_scheme_id"`
}

// OUUpdatable for: PATCH /api/v3/ou/{id}
type OUUpdatable struct {
	Description        string `json:"description"`
	Name               string `json:"name"`
	PermissionSchemeID int    `json:"permission_scheme_id"`
}
