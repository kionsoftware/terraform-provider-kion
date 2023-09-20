package ctclient

// Creation .
type Creation struct {
	RecordID int `json:"record_id"`
	Status   int `json:"status"`
}

// ObjectWithID -
type ObjectWithID struct {
	ID int `json:"id"`
}

// OwnerUser -
type OwnerUser struct {
	ID int `json:"id"`
	// FirstName           string      `json:"first_name"`
	// LastName            string      `json:"last_name"`
	// DisplayName         string      `json:"display_name"`
	// Username            string      `json:"username"`
	// Email               string      `json:"email"`
	// Phone               string      `json:"phone"`
	// IdmsID              int         `json:"idms_id"`
	// PasswordNeedsUpdate bool        `json:"password_needs_update"`
	// Enabled             bool        `json:"enabled"`
	// LastLogin           string      `json:"last_login"`
	// CreatedAt           string      `json:"created_at"`
	// DeletedAt           interface{} `json:"deleted_at"`
}

// OwnerUserGroup -
type OwnerUserGroup struct {
	ID int `json:"id"`
	// Name        string `json:"name"`
	// Description string `json:"description"`
	// IdmsID      int    `json:"idms_id"`
	// Enabled     bool   `json:"enabled"`
	// CreatedAt   string `json:"created_at"`
}

// ChangeOwners -
type ChangeOwners struct {
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
}

type Tag struct {
	Key   string `json:"tag_key"`
	Value string `json:"tag_value"`
}

// AppLabelIdsCreate for: PUT /api/v1/{account|cloud-rule|funding-source|ou|project}/{id}/label
type AppLabelIdsCreate struct {
	LabelIDs []int `json:"app_label_ids"`
}

type AppLabelIdsResponse struct {
	Data []struct {
		ID       int `json:"id"`
		AppLabel struct {
			ID      int    `json:"id"`
			KeyID   int    `json:"key_id"`
			Key     string `json:"key"`
			ValueID int    `json:"value_id"`
			Value   string `json:"value"`
			Color   string `json:"color"`
		} `json:"app_label"`
	} `json:"data"`
	Status int `json:"status"`
}

type ResourceLabelsResponse struct {
	Data []struct {
		Color string `json:"color"`
		ID    int    `json:"id"`
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"data"`
	Status int `json:"status"`
}
