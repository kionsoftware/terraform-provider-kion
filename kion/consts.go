package kion

// Strings repeated across many resources and data sources.
const (
	// Descriptions for the filter block every data source exposes.
	descFilterName   = "The field name whose values you wish to filter by."
	descFilterValues = "The values of the field name you specified."
	descFilterRegex  = "Dictates if the values provided should be treated as regular expressions."

	// Description for the list attribute every data source exposes.
	descDataSourceList = "This is where Kion makes the discovered data available as a list of resources."

	// Description shared by the owner_users and owner_user_groups attributes of
	// every resource that requires an owner.
	descOwnerRequired = "Must provide at least the owner_user_groups field or the owner_users field."

	// Diagnostic summary used wherever last_updated is written back to state.
	summaryLastUpdatedFailed = "Failed to set last_updated"
)
