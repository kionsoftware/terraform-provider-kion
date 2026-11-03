package crud

// templateSDKPaths is the route each verbatim template means by every SDK
// method it names literally. TestTemplateSDKCallsResolveToIntendedPaths checks
// each against the current SDK, since ogen can move a name to another route.
// Prefer (.Op "METHOD /path") in a companion template, which resolves by route.
var templateSDKPaths = map[string]map[string]string{
	"accounthelper/convert.gtpl": {
		"PostAccountCacheConvert": "POST /v3/account-cache/{id}/convert/{project_id}",
		"PostAccountMove":         "POST /v3/account/{id}/move",
		"RevertAccount":           "DELETE /v3/account/revert/{id}",
	},
	"accounthelper/delete.gtpl": {
		"DeleteAccount":      "DELETE /v3/account/{id}",
		"DeleteAccountCache": "DELETE /v3/account-cache/{id}",
	},
	"accounthelper/read.gtpl": {
		"GetAccount":      "GET /v3/account/{id}",
		"GetAccountCache": "GET /v3/account-cache/{id}",
	},
	"accounthelper/status.gtpl": {
		"GetAccountCacheStatus": "GET /v3/account-cache/{id}/status",
	},
	"cloud_account.gtpl": {
		"DeleteAccount":     "DELETE /v3/account/{id}",
		"PatchAccount":      "PATCH /v3/account/{id}",
		"PatchAccountCache": "PATCH /v3/account-cache/{id}",
	},
	"cloud_account_ds_account.gtpl": {
		"GetAccountIndex": "GET /v3/account",
	},
	"cloud_account_ds_aws.gtpl": {
		"GetAccount": "GET /v3/account/{id}",
	},
	"companion_account_linkage_sweep.gtpl": {
		"GetAllAccountLinkages": "GET /v3/account-linkages",
		"RemoveAccountLinkage":  "DELETE /v3/account-linkages/{id}",
	},
	"companion_user_test.gtpl": {
		"GetUser": "GET /v3/user/{id}",
	},
	"companion_webhook_test.gtpl": {
		"GetWebhook": "GET /v3/webhook/{id}",
	},
	"cvoverride.gtpl": {
		"CreateOrUpdateAccountCustomVariable": "PUT /v3/account/{account_id}/custom-variable/{custom_variable_id}",
		"CreateOrUpdateOUCustomVariable":      "PUT /v3/ou/{ou_id}/custom-variable/{custom_variable_id}",
		"CreateOrUpdateProjectCustomVariable": "PUT /v3/project/{project_id}/custom-variable/{custom_variable_id}",
		"DeleteAccountCustomVariable":         "DELETE /v3/account/{account_id}/custom-variable/{custom_variable_id}",
		"DeleteOUCustomVariable":              "DELETE /v3/ou/{ou_id}/custom-variable/{custom_variable_id}",
		"DeleteProjectCustomVariable":         "DELETE /v3/project/{project_id}/custom-variable/{custom_variable_id}",
		"GetAccountCustomVariable":            "GET /v3/account/{account_id}/custom-variable/{custom_variable_id}",
		"GetOUCustomVariable":                 "GET /v3/ou/{ou_id}/custom-variable/{custom_variable_id}",
		"GetProjectCustomVariable":            "GET /v3/project/{project_id}/custom-variable/{custom_variable_id}",
	},
	"cvoverride_ds.gtpl": {
		"GetAccountCustomVariablesByAccountID": "GET /v3/account/{account_id}/custom-variable",
		"GetAccountIndex":                      "GET /v3/account",
		"GetOUCustomVariablesByOUID":           "GET /v3/ou/{ou_id}/custom-variable",
		"GetOUIndex":                           "GET /v3/ou",
		"GetProjectCustomVariablesByProjectID": "GET /v3/project/{project_id}/custom-variable",
		"GetProjectIndex":                      "GET /v3/project",
	},
	"funding_source_permission_mapping_ds.gtpl": {
		"GetFundingSourceUserMappingIndex": "GET /v3/funding-source/{id}/permission-mapping",
	},
	"funding_source_read.gtpl": {
		"GetFundingSourceUserMappingIndex": "GET /v3/funding-source/{id}/permission-mapping",
	},
	"gcp_regions_ds.gtpl": {
		"GetGCPRegionsList": "GET /v3/gcp-resources/list-regions",
	},
	"global_permission_mapping_ds.gtpl": {
		"GetGlobalUserMappingIndex": "GET /v3/global/permission-mapping",
	},
	"ou_permission_mapping_ds.gtpl": {
		"GetOUUserMappingIndex": "GET /v3/ou/{id}/permission-mapping",
	},
	"project.gtpl": {
		"DeleteProject":              "DELETE /v3/project/{id}",
		"GetProject":                 "GET /v3/project/{id}",
		"GetProjectUserMappingIndex": "GET /v3/project/{id}/permission-mapping",
		"PatchProject":               "PATCH /v3/project/{id}",
		"PostProjectMove":            "POST /v3/project/{id}/move",
		"PostProjectWithBudgets":     "POST /v3/project/with-budget",
		"PostProjectWithSpendPlan":   "POST /v3/project/with-spend-plan",
		"UpdateProjectOwners":        "POST /v3/project/{id}/owner",
	},
	"project_enforcement_ds.gtpl": {
		"GetProjectEnforcements": "GET /v3/project/{id}/enforcement",
	},
	"project_note_ds.gtpl": {
		"GetProjectNoteIndex": "GET /v3/project-note",
	},
	"project_permission_mapping_ds.gtpl": {
		"GetProjectUserMappingIndex": "GET /v3/project/{id}/permission-mapping",
	},
	"singleton.gtpl": {
		"GetAppConfigKeyIndex": "GET /v3/app-config",
		"PatchAppConfig":       "PATCH /v3/app-config",
	},
	"singleton_ds.gtpl": {
		"GetAppConfigKeyIndex": "GET /v3/app-config",
	},
}
