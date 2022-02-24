package ctclient

// ProjectFundingCreate for: POST /v3/project
type ProjectFundingCreate struct {
	FundingSourceID int     `json:"funding_source_id"`
	Amount          float64 `json:"amount"`
	StartDatecode   string  `json:"start_datecode"`
	EndDatecode     string  `json:"end_datecode"`
	FundingOrder    int     `json:"funding_order"`
}
