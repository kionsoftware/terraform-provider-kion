package {{ .ServicePackage }}_test

{{ if .IncludeComments -}}
// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the kgen tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the Kion API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.
{{- end }}

import (
{{- if .IncludeComments }}
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
{{- end }}
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"terraform-provider-kion/internal/acctest"
)
{{ if .IncludeComments }}
// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test for a data source.
// We prefix its name with "TestAcc" and the data source name.
//
// Acceptance tests access the Kion API and cost real operations.
{{- end }}
func TestAccKion{{ .DataSource }}DataSource_basic(t *testing.T) {
	{{- if .IncludeComments }}
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	{{- end }}
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.{{ .ProviderResourceName }}.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t) },
		ProtoV6ProviderFactories: acctest.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAcc{{ .DataSource }}DataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					{{- if .IncludeComments }}
					// TIP: Add more checks for specific attributes here.
					// resource.TestCheckResourceAttr(dataSourceName, "name", "test-value"),
					{{- end }}
				),
			},
		},
	})
}

func testAcc{{ .DataSource }}DataSourceConfig_basic() string {
	return `
data "{{ .ProviderResourceName }}" "test" {
  # TIP: Fill in filter criteria or ID to look up the data source.
}
`
}
