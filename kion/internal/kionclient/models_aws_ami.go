package kionclient

// AmiListResponse for: GET /v3/ami
type AmiListResponse struct {
	Data []struct {
		Ami struct {
			AccountID               int        `json:"account_id"`
			AwsAmiID                string     `json:"aws_ami_id"`
			Description             string     `json:"description"`
			ExpiresAt               NullTime   `json:"expires_at"`
			ID                      int        `json:"id"`
			Name                    string     `json:"name"`
			Region                  string     `json:"region"`
			SyncDeprecation         bool       `json:"sync_deprecation"`
			SyncTags                bool       `json:"sync_tags"`
			UnavailableInAws        bool       `json:"unavailable_in_aws"`
			ExpirationAlertNumber   int        `json:"expiration_alert_number"`
			ExpirationAlertUnit     NullString `json:"expiration_alert_unit"`
			ExpirationNotify        bool       `json:"expiration_notify"`
			ExpirationWarningNumber int        `json:"expiration_warning_number"`
			ExpirationWarningUnit   NullString `json:"expiration_warning_unit"`
		} `json:"ami"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AmiResponse for: GET /v3/ami/{id}
type AmiResponse struct {
	Data struct {
		Ami struct {
			AccountID               int        `json:"account_id"`
			AwsAmiID                string     `json:"aws_ami_id"`
			Description             string     `json:"description"`
			ExpiresAt               NullTime   `json:"expires_at"`
			ID                      int        `json:"id"`
			Name                    string     `json:"name"`
			Region                  string     `json:"region"`
			SyncDeprecation         bool       `json:"sync_deprecation"`
			SyncTags                bool       `json:"sync_tags"`
			UnavailableInAws        bool       `json:"unavailable_in_aws"`
			ExpirationAlertNumber   int        `json:"expiration_alert_number"`
			ExpirationAlertUnit     NullString `json:"expiration_alert_unit"`
			ExpirationNotify        bool       `json:"expiration_notify"`
			ExpirationWarningNumber int        `json:"expiration_warning_number"`
			ExpirationWarningUnit   NullString `json:"expiration_warning_unit"`
		} `json:"ami"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AmiCreate for: POST /v3/ami
type AmiCreate struct {
	AccountID               int        `json:"account_id"`
	AwsAmiID                string     `json:"aws_ami_id"`
	Description             string     `json:"description"`
	Name                    string     `json:"name"`
	Region                  string     `json:"region"`
	ExpiresAt               NullTime   `json:"expires_at"`
	SyncDeprecation         bool       `json:"sync_deprecation"`
	SyncTags                bool       `json:"sync_tags"`
	OwnerUserGroupIds       *[]int     `json:"owner_user_group_ids"`
	OwnerUserIds            *[]int     `json:"owner_user_ids"`
	ExpirationAlertNumber   int        `json:"expiration_alert_number"`
	ExpirationAlertUnit     NullString `json:"expiration_alert_unit"`
	ExpirationNotify        bool       `json:"expiration_notify"`
	ExpirationWarningNumber int        `json:"expiration_warning_number"`
	ExpirationWarningUnit   NullString `json:"expiration_warning_unit"`
}

// AmiUpdate for: PATCH /v3/ami/{id}
type AmiUpdate struct {
	Description             string     `json:"description"`
	Name                    string     `json:"name"`
	Region                  string     `json:"region"`
	SyncDeprecation         bool       `json:"sync_deprecation"`
	SyncTags                bool       `json:"sync_tags"`
	ExpiresAt               NullTime   `json:"expires_at"`
	ExpirationAlertNumber   int        `json:"expiration_alert_number"`
	ExpirationAlertUnit     NullString `json:"expiration_alert_unit"`
	ExpirationNotify        bool       `json:"expiration_notify"`
	ExpirationWarningNumber int        `json:"expiration_warning_number"`
	ExpirationWarningUnit   NullString `json:"expiration_warning_unit"`
}
