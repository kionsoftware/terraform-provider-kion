package crud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// kion_account_cache with an account_number must import that account, never
// create a new one. The create route makes a real AWS account through
// Organizations, which is what every apply did before raw_create learned the
// import route.
func TestAccountCacheImportsAnExistingAccount(t *testing.T) {
	t.Parallel()
	src, err := os.ReadFile(filepath.Join("..", "..", "service", "account_cache", "account_cache.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`importing := !plan.AccountNumber.IsNull() && !plan.AccountNumber.IsUnknown() && plan.AccountNumber.ValueString() != ""`,
		`RawPost(ctx, conns.PathUnless(importing, "/v3/account-cache/create?account-type=aws", "/v3/account-cache?account-type=aws"), rawBody)`,
	} {
		if !strings.Contains(string(src), want) {
			t.Errorf("account_cache Create lacks %q; run `make crud-force`", want)
		}
	}
}
