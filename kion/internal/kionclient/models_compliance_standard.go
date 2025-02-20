package kionclient

// ComplianceStandardListResponse for: GET /api/v3/compliance/standard
type ComplianceStandardListResponse struct {
	Data []struct {
		CreatedAt       string `json:"created_at"`
		CreatedByUserID int    `json:"created_by_user_id"`
		CtManaged       bool   `json:"ct_managed"`
		Description     string `json:"description"`
		ID              int    `json:"id"`
		Name            string `json:"name"`
	} `json:"data"`
	Status int `json:"status"`
}

// ComplianceStandardResponse for: GET /api/v3/compliance/standard/{id}
type ComplianceStandardResponse struct {
	Data struct {
		ComplianceChecks   []ObjectWithID `json:"compliance_checks"`
		ComplianceStandard struct {
			CreatedAt       string `json:"created_at"`
			CreatedByUserID int    `json:"created_by_user_id"`
			CtManaged       bool   `json:"ct_managed"`
			Description     string `json:"description"`
			ID              int    `json:"id"`
			Name            string `json:"name"`
		} `json:"compliance_standard"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// ComplianceStandardCreate for: POST /api/v3/compliance/standard
type ComplianceStandardCreate struct {
	ComplianceCheckIds *[]int `json:"compliance_check_ids"`
	CreatedByUserID    int    `json:"created_by_user_id"`
	Description        string `json:"description"`
	Name               string `json:"name"`
	OwnerUserGroupIds  *[]int `json:"owner_user_group_ids"`
	OwnerUserIds       *[]int `json:"owner_user_ids"`
}

// ComplianceStandardUpdate for: PATCH /api/v3/compliance/standard/{id}
type ComplianceStandardUpdate struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

// ComplianceStandardAssociationsAdd for: POST /api/v3/compliance/standard/{id}/association
type ComplianceStandardAssociationsAdd struct {
	ComplianceCheckIds *[]int `json:"compliance_check_ids"`
}

// ComplianceStandardAssociationsRemove for: DELETE /v3/compliance/standard/{id}/association
type ComplianceStandardAssociationsRemove struct {
	ComplianceCheckIds []int `json:"compliance_check_ids"`
}
