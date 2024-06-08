package kionclient

// AzurePolicyListResponse for: GET /api/v3/azure-policy
type AzurePolicyListResponse struct {
	Data []struct {
		AzurePolicy struct {
			AzureManagedPolicyDefID string `json:"azure_managed_policy_def_id"`
			CtManaged               bool   `json:"ct_managed"`
			Description             string `json:"description"`
			ID                      int    `json:"id"`
			Name                    string `json:"name"`
			Parameters              string `json:"parameters"`
			Policy                  string `json:"policy"`
		} `json:"azure_policy"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzurePolicyResponse for: GET /api/v3/azure-policy/{id}
type AzurePolicyResponse struct {
	Data struct {
		AzurePolicy struct {
			AzureManagedPolicyDefID string `json:"azure_managed_policy_def_id"`
			CtManaged               bool   `json:"ct_managed"`
			Description             string `json:"description"`
			ID                      int    `json:"id"`
			Name                    string `json:"name"`
			Parameters              string `json:"parameters"`
			Policy                  string `json:"policy"`
		} `json:"azure_policy"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzurePolicy for AzurePolicyCreate
type AzurePolicy struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Parameters  string `json:"parameters"`
	Policy      string `json:"policy"`
}

// AzurePolicyCreate for: POST /api/v3/azure-policy
type AzurePolicyCreate struct {
	AzurePolicy     AzurePolicy `json:"azure_policy"`
	OwnerUserGroups *[]int      `json:"owner_user_groups"`
	OwnerUsers      *[]int      `json:"owner_users"`
}

// AzurePolicyDefinitionUpdate for: PATCH /api/v3/azure-policy/{id}
type AzurePolicyDefinitionUpdate struct {
	Description string `json:"description"`
	Name        string `json:"name"`
	Parameters  string `json:"parameters"`
	Policy      string `json:"policy"`
}
