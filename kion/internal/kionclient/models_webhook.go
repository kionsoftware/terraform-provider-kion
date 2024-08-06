package kionclient

// Webhook represents a webhook configuration in the system.
type Webhook struct {
	ID                   int    `json:"id"`
	CalloutURL           string `json:"callout_url"`
	Description          string `json:"description"`
	Name                 string `json:"name"`
	RequestBody          string `json:"request_body,omitempty"`
	RequestHeaders       string `json:"request_headers,omitempty"`
	RequestMethod        string `json:"request_method"`
	ShouldSendSecureInfo bool   `json:"should_send_secure_info,omitempty"`
	SkipSSL              bool   `json:"skip_ssl,omitempty"`
	TimeoutInSeconds     int    `json:"timeout_in_seconds,omitempty"`
	UseRequestHeaders    bool   `json:"use_request_headers,omitempty"`
	OwnerUserIDs         []int  `json:"owner_user_ids,omitempty"`
	OwnerUserGroupIDs    []int  `json:"owner_user_group_ids,omitempty"`
}

// WebhookCreateResponse represents the response from the API for a created webhook.
type WebhookCreateResponse struct {
	RecordID int `json:"record_id"`
	Status   int `json:"status"`
}

// WebhookWithOwnersResponse represents the response for a single webhook with owner details.
type WebhookWithOwnersResponse struct {
	Data struct {
		Webhook         Webhook          `json:"webhook"`
		OwnerUserGroups []OwnerUserGroup `json:"owner_user_groups"`
		OwnerUsers      []OwnerUser      `json:"owner_users"`
	} `json:"data"`
	Status int `json:"status"`
}
