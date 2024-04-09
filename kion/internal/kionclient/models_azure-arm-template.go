package kionclient

// AzureARMTemplateListResponse for: GET /api/v3/azure-arm-template
type AzureARMTemplateListResponse struct {
	Data []struct {
		AzureArmTemplate struct {
			CtManaged             bool   `json:"ct_managed"`
			DeploymentMode        int    `json:"deployment_mode"`
			Description           string `json:"description"`
			ID                    int    `json:"id"`
			Name                  string `json:"name"`
			ResourceGroupName     string `json:"resource_group_name"`
			ResourceGroupRegionID int    `json:"resource_group_region_id"`
			Template              string `json:"template"`
			TemplateParameters    string `json:"template_parameters"`
			Version               int    `json:"version"`
		} `json:"azure_arm_template"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzureARMTemplateResponse for: GET /api/v3/azure-arm-template/{id}
type AzureARMTemplateResponse struct {
	Data struct {
		AzureArmTemplate struct {
			CtManaged             bool   `json:"ct_managed"`
			DeploymentMode        int    `json:"deployment_mode"`
			Description           string `json:"description"`
			ID                    int    `json:"id"`
			Name                  string `json:"name"`
			ResourceGroupName     string `json:"resource_group_name"`
			ResourceGroupRegionID int    `json:"resource_group_region_id"`
			Template              string `json:"template"`
			TemplateParameters    string `json:"template_parameters"`
			Version               int    `json:"version"`
		} `json:"azure_arm_template"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AzureARMTemplateDefinitionCreate for: POST /api/v3/azure-arm-template
type AzureARMTemplateDefinitionCreate struct {
	DeploymentMode        int    `json:"deployment_mode"`
	Description           string `json:"description"`
	Name                  string `json:"name"`
	OwnerUserGroupIds     *[]int `json:"owner_user_group_ids"`
	OwnerUserIds          *[]int `json:"owner_user_ids"`
	ResourceGroupName     string `json:"resource_group_name"`
	ResourceGroupRegionID int    `json:"resource_group_region_id"`
	Template              string `json:"template"`
	TemplateParameters    string `json:"template_parameters"`
}

// AzureARMTemplateDefinitionUpdate for: PATCH /api/v3/azure-arm-template/{id}
type AzureARMTemplateDefinitionUpdate struct {
	DeploymentMode     int    `json:"deployment_mode"`
	Description        string `json:"description"`
	Name               string `json:"name"`
	Template           string `json:"template"`
	TemplateParameters string `json:"template_parameters"`
}
