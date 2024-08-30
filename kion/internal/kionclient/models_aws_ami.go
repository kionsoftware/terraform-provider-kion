package kionclient

import "time"

// AmiListResponse for: GET /v3/ami
type AmiListResponse struct {
	Data []struct {
		Ami struct {
			AccountID        int       `json:"account_id"`
			AwsAmiID         string    `json:"aws_ami_id"`
			Description      string    `json:"description"`
			ExpiresAt        time.Time `json:"expires_at"`
			ID               int       `json:"id"`
			Name             string    `json:"name"`
			Region           string    `json:"region"`
			SyncDeprecation  bool      `json:"sync_deprecation"`
			SyncTags         bool      `json:"sync_tags"`
			UnavailableInAws bool      `json:"unavailable_in_aws"`
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
			AccountID        int       `json:"account_id"`
			AwsAmiID         string    `json:"aws_ami_id"`
			Description      string    `json:"description"`
			ExpiresAt        time.Time `json:"expires_at"`
			ID               int       `json:"id"`
			Name             string    `json:"name"`
			Region           string    `json:"region"`
			SyncDeprecation  bool      `json:"sync_deprecation"`
			SyncTags         bool      `json:"sync_tags"`
			UnavailableInAws bool      `json:"unavailable_in_aws"`
		} `json:"ami"`
		OwnerUserGroups []ObjectWithID `json:"owner_user_groups"`
		OwnerUsers      []ObjectWithID `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}

// AmiCreate for: POST /v3/ami
type AmiCreate struct {
	AccountID         int    `json:"account_id"`
	AwsAmiID          string `json:"aws_ami_id"`
	Description       string `json:"description"`
	Name              string `json:"name"`
	Region            string `json:"region"`
	ExpiresAt         string `json:"expires_at"`
	SyncDeprecation   bool   `json:"sync_deprecation"`
	SyncTags          bool   `json:"sync_tags"`
	OwnerUserGroupIds *[]int `json:"owner_user_group_ids"`
	OwnerUserIds      *[]int `json:"owner_user_ids"`
}

// AmiUpdate for: PATCH /v3/ami/{id}
type AmiUpdate struct {
	Description     string `json:"description"`
	Name            string `json:"name"`
	Region          string `json:"region"`
	SyncDeprecation bool   `json:"sync_deprecation"`
	SyncTags        bool   `json:"sync_tags"`
	ExpiresAt       string `json:"expires_at"`
}
