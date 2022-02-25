package ctclient

// GroupAssociationListResponse for: GET /api/v3/idms/{id}/group-association
type GroupAssociationListResponse struct {
	Data []struct {
		AssertionName       string `json:"assertion_name"`
		AssertionRegex      string `json:"assertion_regex"`
		ID                  int    `json:"id"`
		IdmsID              int    `json:"idms_id"`
		IdmsSamlID          int    `json:"idms_saml_id"`
		ShouldUpdateOnLogin bool   `json:"should_update_on_login"`
		UserGroupID         int    `json:"user_group_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// GroupAssociationResponse for: GET /api/v3/idms/group-association/{id}
type GroupAssociationResponse struct {
	Data struct {
		AssertionName       string `json:"assertion_name"`
		AssertionRegex      string `json:"assertion_regex"`
		ID                  int    `json:"id"`
		IdmsID              int    `json:"idms_id"`
		IdmsSamlID          int    `json:"idms_saml_id"`
		ShouldUpdateOnLogin bool   `json:"should_update_on_login"`
		UserGroupID         int    `json:"user_group_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// CreateSAMLGroupAssociation for: POST /api/v3/idms/group-association
type CreateSAMLGroupAssociation struct {
	AssertionName  string `json:"assertion_name"`
	AssertionRegex string `json:"assertion_regex"`
	IdmsID         int    `json:"idms_id"`
	UpdateOnLogin  bool   `json:"update_on_login"`
	UserGroupID    int    `json:"user_group_id"`
}

// UpdateSAMLGroupAssociation for: PATCH /api/v3/idms/group-association/{id}
type UpdateSAMLGroupAssociation struct {
	AssertionName  string `json:"assertion_name"`
	AssertionRegex string `json:"assertion_regex"`
	UpdateOnLogin  bool   `json:"update_on_login"`
	UserGroupID    int    `json:"user_group_id"`
}
