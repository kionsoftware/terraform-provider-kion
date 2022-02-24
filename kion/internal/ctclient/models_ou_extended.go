package ctclient

// OUPermissionAdd for: POST /v3/ou/{id}/permission-mapping
type OUPermissionAdd struct {
	AppRoleID         *int   `json:"app_role_id"`
	OwnerUserGroupIds *[]int `json:"user_groups_ids"`
	OwnerUserIds      *[]int `json:"user_ids"`
	PostWebhookID     *int   `json:"post_webhook_id"`
	PreWebhookID      *int   `json:"pre_webhook_id"`
}
