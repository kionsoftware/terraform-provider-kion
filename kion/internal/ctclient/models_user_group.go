package ctclient

// UGroupListResponse for: GET /api/v3/user-group
type UGroupListResponse struct {
	Data []struct {
		CreatedAt   string `json:"created_at"`
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		ID          int    `json:"id"`
		IdmsID      int    `json:"idms_id"`
		Name        string `json:"name"`
	} `json:"data"`
	Status int `json:"status"`
}

// UGroupResponse for: GET /api/v3/user-group/{id}
type UGroupResponse struct {
	Data struct {
		OwnerGroup []ObjectWithID `json:"owner_group"`
		OwnerUsers []ObjectWithID `json:"owner_users"`
		UserGroup  struct {
			CreatedAt   string `json:"created_at"`
			Description string `json:"description"`
			Enabled     bool   `json:"enabled"`
			ID          int    `json:"id"`
			IdmsID      int    `json:"idms_id"`
			Name        string `json:"name"`
		} `json:"user_group"`
		Users []ObjectWithID `json:"users"`
	} `json:"data"`
	Status int `json:"status"`
}

// UGroupCreate for: POST /api/v3/user-group
type UGroupCreate struct {
	Description       string `json:"description"`
	IdmsID            int    `json:"idms_id"`
	Name              string `json:"name"`
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
	UserIds           *[]int `json:"user_ids"`
}

// UGroupUpdatable for: PATCH /api/v3/user-group/{id}
type UGroupUpdatable struct {
	Description string `json:"description"`
	IdmsID      int    `json:"idms_id"`
	Name        string `json:"name"`
}

// UserGroupAssociationsAdd for: POST /api/v3/user-group/{id}/user
type UserGroupAssociationsAdd []int

// UserGroupAssociationsRemove for: DELETE /api/v3/user-group/{id}/user
type UserGroupAssociationsRemove []int
