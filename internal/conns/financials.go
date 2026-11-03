package conns

import (
	"context"
	"encoding/json"
	"fmt"
)

// appConfigResponse is the slice of GET /v3/app-config this package needs.
type appConfigResponse struct {
	Data struct {
		BudgetMode bool `json:"budget_mode"`
	} `json:"data"`
}

// DetectFinancialMode queries GET {APIURL}/v3/app-config and records whether the
// install funds projects with budgets or with spend plans.
//
// The mode is not cosmetic: POST /v3/project rejects a budget-mode install
// before it parses the body, so kion_project has to choose between
// /v3/project/with-budget and /v3/project/with-spend-plan before it sends
// anything, and no payload can make the wrong one work.
//
// Best-effort like DetectVersion: reading app-config needs a global settings
// permission that a token which may only create projects need not hold, so a
// failure leaves the mode undetected rather than failing Configure. Create then
// takes the mode from whichever of budget / project_funding the configuration
// supplies, and says so if it supplies neither.
func (c *KionClient) DetectFinancialMode(ctx context.Context) error {
	body, err := c.RawGet(ctx, "/v3/app-config")
	if err != nil {
		return fmt.Errorf("querying app-config: %w", err)
	}

	var ac appConfigResponse
	if err := json.Unmarshal(body, &ac); err != nil {
		return fmt.Errorf("decoding app-config response: %w", err)
	}

	c.BudgetMode = ac.Data.BudgetMode
	c.FinancialModeDetected = true
	return nil
}
