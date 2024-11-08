package kionclient

// CustomVariableListResponse for: GET /api/v3/compliance/custom-variable
type CustomVariableListResponse struct {
	Data struct {
		Items []struct {
			ID                     int         `json:"id"`
			Name                   string      `json:"name"`
			Description            string      `json:"description"`
			Type                   string      `json:"type"`
			DefaultValue           interface{} `json:"default_value"`
			ValueValidationRegex   string      `json:"value_validation_regex"`
			ValueValidationMessage string      `json:"value_validation_message"`
			KeyValidationRegex     string      `json:"key_validation_regex"`
			KeyValidationMessage   string      `json:"key_validation_message"`
			OwnerUserIDs           []int       `json:"owner_user_ids"`
			OwnerUserGroupIDs      []int       `json:"owner_user_group_ids"`
		} `json:"items"`
	} `json:"data"`
	Status int `json:"status"`
}

// CustomVariableResponse for: GET /api/v3/custom-variable/{id}
type CustomVariableResponse struct {
	Data struct {
		ID                     int         `json:"id"`
		Name                   string      `json:"name"`
		Description            string      `json:"description"`
		Type                   string      `json:"type"`
		DefaultValue           interface{} `json:"default_value"`
		ValueValidationRegex   string      `json:"value_validation_regex"`
		ValueValidationMessage string      `json:"value_validation_message"`
		KeyValidationRegex     string      `json:"key_validation_regex"`
		KeyValidationMessage   string      `json:"key_validation_message"`
		OwnerUserIDs           []int       `json:"owner_user_ids"`
		OwnerUserGroupIDs      []int       `json:"owner_user_group_ids"`
	} `json:"data"`
	Status int `json:"status"`
}

// CustomVariableCreate for: POST /api/v3/custom-variable
type CustomVariableCreate struct {
	Name                   string      `json:"name"`
	Description            string      `json:"description"`
	Type                   string      `json:"type"`
	DefaultValue           interface{} `json:"default_value"`
	ValueValidationRegex   string      `json:"value_validation_regex"`
	ValueValidationMessage string      `json:"value_validation_message"`
	KeyValidationRegex     string      `json:"key_validation_regex"`
	KeyValidationMessage   string      `json:"key_validation_message"`
	OwnerUserIDs           []int       `json:"owner_user_ids"`
	OwnerUserGroupIDs      []int       `json:"owner_user_group_ids"`
}

// CustomVariableUpdate for: PATCH /api/v3/custom-variable/{id}
type CustomVariableUpdate struct {
	Description            string      `json:"description"`
	DefaultValue           interface{} `json:"default_value"`
	ValueValidationRegex   string      `json:"value_validation_regex"`
	ValueValidationMessage string      `json:"value_validation_message"`
	KeyValidationRegex     string      `json:"key_validation_regex"`
	KeyValidationMessage   string      `json:"key_validation_message"`
	OwnerUserIDs           []int       `json:"owner_user_ids"`
	OwnerUserGroupIDs      []int       `json:"owner_user_group_ids"`
}
