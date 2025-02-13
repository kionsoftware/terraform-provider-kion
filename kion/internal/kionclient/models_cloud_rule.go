package kionclient

// CloudRuleListResponse for: GET /api/v3/cloud-rule
type CloudRuleListResponse struct {
	Data []struct {
		BuiltIn           bool   `json:"built_in"`
		ConcurrentCftSync bool   `json:"concurrent_cft_sync"`
		Description       string `json:"description"`
		ID                int    `json:"id"`
		Name              string `json:"name"`
		PostWebhookID     *int   `json:"post_webhook_id"`
		PreWebhookID      *int   `json:"pre_webhook_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// CloudRuleResponse for: GET /api/v3/cloud-rule/{id}
type CloudRuleResponse struct {
	Data struct {
		AwsCloudformationTemplates  []ObjectWithID `json:"aws_cloudformation_templates"`
		AwsIamPolicies              []ObjectWithID `json:"aws_iam_policies"`
		AzureArmTemplateDefinitions []ObjectWithID `json:"azure_arm_template_definitions"`
		AzurePolicyDefinitions      []ObjectWithID `json:"azure_policy_definitions"`
		AzureRoleDefinitions        []ObjectWithID `json:"azure_role_definitions"`
		CloudRule                   struct {
			BuiltIn           bool   `json:"built_in"`
			ConcurrentCftSync bool   `json:"concurrent_cft_sync"`
			Description       string `json:"description"`
			ID                int    `json:"id"`
			Name              string `json:"name"`
			PostWebhookID     *int   `json:"post_webhook_id"`
			PreWebhookID      *int   `json:"pre_webhook_id"`
		} `json:"cloud_rule"`
		ComplianceStandards                 []ObjectWithID `json:"compliance_standards"`
		GCPIAMRoles                         []ObjectWithID `json:"gcp_iam_roles"`
		InternalAwsAmis                     []ObjectWithID `json:"internal_aws_amis"`
		InternalAwsServiceCatalogPortfolios []ObjectWithID `json:"internal_aws_service_catalog_portfolios"`
		OUs                                 []ObjectWithID `json:"ous"`
		OwnerUserGroups                     []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers                          []ObjectWithID `json:"owner_users"`
		Projects                            []ObjectWithID `json:"projects"`
		ServiceControlPolicies              []ObjectWithID `json:"service_control_policies"`
	} `json:"data"`
	Status int `json:"status"`
}

// CloudRuleCreate for: POST /api/v3/cloud-rule
type CloudRuleCreate struct {
	AzureArmTemplateDefinitionIds *[]int `json:"azure_arm_template_definition_ids"`
	AzurePolicyDefinitionIds      *[]int `json:"azure_policy_definition_ids"`
	AzureRoleDefinitionIds        *[]int `json:"azure_role_definition_ids"`
	CftIds                        *[]int `json:"cft_ids"`
	ComplianceStandardIds         *[]int `json:"compliance_standard_ids"`
	ConcurrentCftSync             bool   `json:"concurrent_cft_sync"`
	Description                   string `json:"description"`
	GcpIamRoleIds                 *[]int `json:"gcp_iam_role_ids"`
	IamPolicyIds                  *[]int `json:"iam_policy_ids"`
	InternalAmiIds                *[]int `json:"internal_ami_ids"`
	InternalPortfolioIds          *[]int `json:"internal_portfolio_ids"`
	Name                          string `json:"name"`
	OUIds                         *[]int `json:"ou_ids"`
	OwnerUserGroupIds             *[]int `json:"owner_user_group_ids"`
	OwnerUserIds                  *[]int `json:"owner_user_ids"`
	PostWebhookID                 *int   `json:"post_webhook_id"`
	PreWebhookID                  *int   `json:"pre_webhook_id"`
	ProjectIds                    *[]int `json:"project_ids"`
	ServiceControlPolicyIds       *[]int `json:"service_control_policy_ids"`
}

// CloudRuleUpdate for: PATCH /api/v3/cloud-rule/{id}
type CloudRuleUpdate struct {
	ConcurrentCftSync bool   `json:"concurrent_cft_sync"`
	Description       string `json:"description"`
	Name              string `json:"name"`
	PostWebhookID     *int   `json:"post_webhook_id"`
	PreWebhookID      *int   `json:"pre_webhook_id"`
}

// CloudRuleAssociationsAdd for: POST /api/v3/cloud-rule/{id}/association
type CloudRuleAssociationsAdd struct {
	AzureArmTemplateDefinitionIds *[]int `json:"azure_arm_template_definition_ids"`
	AzurePolicyDefinitionIds      *[]int `json:"azure_policy_definition_ids"`
	AzureRoleDefinitionIds        *[]int `json:"azure_role_definition_ids"`
	CftIds                        *[]int `json:"cft_ids"`
	ComplianceStandardIds         *[]int `json:"compliance_standard_ids"`
	GcpIamRoleIds                 *[]int `json:"gcp_iam_role_ids"`
	IamPolicyIds                  *[]int `json:"iam_policy_ids"`
	InternalAmiIds                *[]int `json:"internal_ami_ids"`
	InternalPortfolioIds          *[]int `json:"internal_portfolio_ids"`
	OUIds                         *[]int `json:"ou_ids"`
	ProjectIds                    *[]int `json:"project_ids"`
	ServiceControlPolicyIds       *[]int `json:"service_control_policy_ids"`
}

// CloudRuleAssociationsRemove for: DELETE /api/v3/cloud-rule/{id}/association
type CloudRuleAssociationsRemove struct {
	AzureArmTemplateDefinitionIds *[]int `json:"azure_arm_template_definition_ids"`
	AzurePolicyDefinitionIds      *[]int `json:"azure_policy_definition_ids"`
	AzureRoleDefinitionIds        *[]int `json:"azure_role_definition_ids"`
	CftIds                        *[]int `json:"cft_ids"`
	ComplianceStandardIds         *[]int `json:"compliance_standard_ids"`
	GcpIamRoleIds                 *[]int `json:"gcp_iam_role_ids"`
	IamPolicyIds                  *[]int `json:"iam_policy_ids"`
	InternalAmiIds                *[]int `json:"internal_ami_ids"`
	InternalPortfolioIds          *[]int `json:"internal_portfolio_ids"`
	OUIds                         *[]int `json:"ou_ids"`
	ProjectIds                    *[]int `json:"project_ids"`
	ServiceControlPolicyIds       *[]int `json:"service_control_policy_ids"`
}
