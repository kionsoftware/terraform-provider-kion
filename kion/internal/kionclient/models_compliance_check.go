package kionclient

// ComplianceCheckListResponse for: GET /api/v3/compliance/check
type ComplianceCheckListResponse struct {
	Data []struct {
		AzurePolicyID         *int     `json:"azure_policy_id"`
		Body                  string   `json:"body"`
		CloudProviderID       int      `json:"cloud_provider_id"`
		ComplianceCheckTypeID int      `json:"compliance_check_type_id"`
		CreatedAt             string   `json:"created_at"`
		CreatedByUserID       int      `json:"created_by_user_id"`
		CtManaged             bool     `json:"ct_managed"`
		Description           string   `json:"description"`
		FrequencyMinutes      int      `json:"frequency_minutes"`
		FrequencyTypeID       int      `json:"frequency_type_id"`
		ID                    int      `json:"id"`
		IsAllRegions          bool     `json:"is_all_regions"`
		IsAutoArchived        bool     `json:"is_auto_archived"`
		LastScanID            int      `json:"last_scan_id"`
		Name                  string   `json:"name"`
		Regions               []string `json:"regions"`
		SeverityTypeID        *int     `json:"severity_type_id"`
	} `json:"data"`
	Status int `json:"status"`
}

// ComplianceCheckWithOwnersResponse for: GET /api/v3/compliance/check/{id}
type ComplianceCheckWithOwnersResponse struct {
	Data struct {
		ComplianceCheck struct {
			AzurePolicyID         *int     `json:"azure_policy_id"`
			Body                  string   `json:"body"`
			CloudProviderID       int      `json:"cloud_provider_id"`
			ComplianceCheckTypeID int      `json:"compliance_check_type_id"`
			CreatedAt             string   `json:"created_at"`
			CreatedByUserID       int      `json:"created_by_user_id"`
			CtManaged             bool     `json:"ct_managed"`
			Description           string   `json:"description"`
			FrequencyMinutes      int      `json:"frequency_minutes"`
			FrequencyTypeID       int      `json:"frequency_type_id"`
			ID                    int      `json:"id"`
			IsAllRegions          bool     `json:"is_all_regions"`
			IsAutoArchived        bool     `json:"is_auto_archived"`
			LastScanID            int      `json:"last_scan_id"`
			Name                  string   `json:"name"`
			Regions               []string `json:"regions"`
			SeverityTypeID        *int     `json:"severity_type_id"`
		} `json:"compliance_check"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// ComplianceCheckCreate for: POST /api/v3/compliance/check
type ComplianceCheckCreate struct {
	AzurePolicyID         *int     `json:"azure_policy_id"`
	Body                  string   `json:"body"`
	CloudProviderID       int      `json:"cloud_provider_id"`
	ComplianceCheckTypeID int      `json:"compliance_check_type_id"`
	CreatedByUserID       int      `json:"created_by_user_id"`
	Description           string   `json:"description"`
	FrequencyMinutes      int      `json:"frequency_minutes"`
	FrequencyTypeID       int      `json:"frequency_type_id"`
	IsAllRegions          bool     `json:"is_all_regions"`
	IsAutoArchived        bool     `json:"is_auto_archived"`
	Name                  string   `json:"name"`
	OwnerUserGroupIds     *[]int   `json:"owner_user_group_ids"`
	OwnerUserIds          *[]int   `json:"owner_user_ids"`
	Regions               []string `json:"regions"`
	SeverityTypeID        *int     `json:"severity_type_id"`
}

// ComplianceCheckUpdate for: PATCH /api/v3/compliance/check/{id}
type ComplianceCheckUpdate struct {
	AzurePolicyID         *int     `json:"azure_policy_id"`
	Body                  string   `json:"body"`
	CloudProviderID       int      `json:"cloud_provider_id"`
	ComplianceCheckTypeID int      `json:"compliance_check_type_id"`
	Description           string   `json:"description"`
	FrequencyMinutes      int      `json:"frequency_minutes"`
	FrequencyTypeID       int      `json:"frequency_type_id"`
	IsAllRegions          bool     `json:"is_all_regions"`
	IsAutoArchived        bool     `json:"is_auto_archived"`
	Name                  string   `json:"name"`
	Regions               []string `json:"regions"`
	SeverityTypeID        *int     `json:"severity_type_id"`
}
