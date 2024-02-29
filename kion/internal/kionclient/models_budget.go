package kionclient

// BudgetCreate for: POST /api/v3/project/with-budget
type BudgetCreate struct {
	ProjectID        int                `json:"project_id"`
	OUID             int                `json:"ou_id"`
	StartDatecode    string             `json:"start_datecode"`
	EndDatecode      string             `json:"end_datecode"`
	Amount           float64            `json:"amount"`
	FundingSourceIDs *[]int             `json:"funding_source_ids"`
	Data             []BudgetDataCreate `json:"data"`
}

// BudgetDataCreate for: POST /api/v3/project/with-budget
type BudgetDataCreate struct {
	Datecode        string  `json:"datecode"`
	Amount          float64 `json:"amount"`
	FundingSourceID int     `json:"funding_source_id"`
	Priority        int     `json:"priority"`
}
