package ctclient

// CFTListResponseWithOwners for: GET /api/v3/cft
type CFTListResponseWithOwners struct {
	Data []struct {
		Cft struct {
			Description           string   `json:"description"`
			ID                    int      `json:"id"`
			Name                  string   `json:"name"`
			Policy                string   `json:"policy"`
			Region                string   `json:"region"`
			Regions               []string `json:"regions"`
			SnsArns               string   `json:"sns_arns"`
			TemplateParameters    string   `json:"template_parameters"`
			TerminationProtection bool     `json:"termination_protection"`
		} `json:"cft"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// CFTResponseWithOwners for: GET /api/v3/cft/{id}
type CFTResponseWithOwners struct {
	Data struct {
		Cft struct {
			Description           string   `json:"description"`
			ID                    int      `json:"id"`
			Name                  string   `json:"name"`
			Policy                string   `json:"policy"`
			Region                string   `json:"region"`
			Regions               []string `json:"regions"`
			SnsArns               string   `json:"sns_arns"`
			TemplateParameters    string   `json:"template_parameters"`
			TerminationProtection bool     `json:"termination_protection"`
		} `json:"cft"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// CFTCreate for: POST /api/v3/cft
type CFTCreate struct {
	Description           string   `json:"description"`
	Name                  string   `json:"name"`
	OwnerUserGroupIds     *[]int   `json:"owner_user_group_ids"`
	OwnerUserIds          *[]int   `json:"owner_user_ids"`
	Policy                string   `json:"policy"`
	Region                string   `json:"region"`
	Regions               []string `json:"regions"`
	SnsArns               string   `json:"sns_arns"`
	TemplateParameters    string   `json:"template_parameters"`
	TerminationProtection bool     `json:"termination_protection"`
}

// CFTUpdate for: PATCH /api/v3/cft/{id}
type CFTUpdate struct {
	Description           string   `json:"description"`
	Name                  string   `json:"name"`
	Policy                string   `json:"policy"`
	Region                string   `json:"region"`
	Regions               []string `json:"regions"`
	SnsArns               string   `json:"sns_arns"`
	TemplateParameters    string   `json:"template_parameters"`
	TerminationProtection bool     `json:"termination_protection"`
}
