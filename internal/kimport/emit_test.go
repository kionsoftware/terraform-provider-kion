package kimport

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func results() []Result {
	return []Result{
		{TFType: "kion_ou", Status: "ok", Records: []Record{
			{ID: "1", Name: "Root"}, {ID: "5", Name: "Central Services"},
		}},
		{TFType: "kion_aws_resource_tag", Status: "unsupported", Reason: "kind: no_read"},
	}
}

func TestRenderImportsEmitsABlockPerRecord(t *testing.T) {
	t.Parallel()
	out := RenderImports(results(), "1.0.0")
	assert.Equal(t, 2, strings.Count(out, "import {"))
	assert.Contains(t, out, `to = kion_ou.root`)
	assert.Contains(t, out, `id = "5"`)
}

func TestRenderImportsIncludesTheProviderRequirement(t *testing.T) {
	t.Parallel()
	out := RenderImports(results(), "1.0.0")
	assert.Contains(t, out, `source  = "kionsoftware/kion"`)
	assert.Contains(t, out, `version = "1.0.0"`)
}

func TestRenderImportsNotesTheGapsAsComments(t *testing.T) {
	t.Parallel()
	out := RenderImports(results(), "1.0.0")
	assert.Contains(t, out, "# kion_aws_resource_tag")
	assert.Contains(t, out, "kind: no_read")
}

func TestRenderImportsIsDeterministic(t *testing.T) {
	t.Parallel()
	assert.Equal(t, RenderImports(results(), "1.0.0"), RenderImports(results(), "1.0.0"))
}

// --- Correction 2: a successful read with skipped records must not render
// as clean. Results with status "ok" (or "empty") but a non-empty Reason
// carry a real gap -- e.g. records skipped for missing id, association
// records skipped for a missing key field, or a flat-list 405 fallback --
// and must not be silently dropped just because Status isn't unsupported/error.

func TestRenderImportsNotesCaveatsForOkResultsWithReason(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_scope", Status: "ok", Records: []Record{{ID: "1", Name: "Prod"}},
			Reason: "2 record(s) skipped: no id"},
	}
	out := RenderImports(rs, "1.0.0")
	assert.Contains(t, out, "# Read with caveats:")
	assert.Contains(t, out, "kion_scope")
	assert.Contains(t, out, "2 record(s) skipped: no id")
}

func TestRenderImportsOmitsCaveatsWhenReasonEmpty(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_ou", Status: "ok", Records: []Record{{ID: "1", Name: "Root"}}, Reason: ""},
	}
	out := RenderImports(rs, "1.0.0")
	assert.NotContains(t, out, "# Read with caveats:")
}

func TestRenderImportsCaveatsDoNotDuplicateNotImportedGaps(t *testing.T) {
	t.Parallel()
	// A result already reported under "# Not imported:" (unsupported/error)
	// must not also appear under "# Read with caveats:".
	out := RenderImports(results(), "1.0.0")
	assert.Contains(t, out, "# Not imported:")
	assert.NotContains(t, out, "# Read with caveats:")
}

// TestRenderImportsSeparatesGapsFromCaveatsWithBlankLine: when both the
// "Not imported:" and "Read with caveats:" blocks render, they must not be
// line-adjacent -- a bare comment header sitting directly between two
// unrelated tables of resources reads as one continuous list. A blank line
// must separate them.
func TestRenderImportsSeparatesGapsFromCaveatsWithBlankLine(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_aws_resource_tag", Status: "unsupported", Reason: "kind: no_read"},
		{TFType: "kion_compliance_family", Status: "ok",
			Records: []Record{{ID: "1", Name: "PCI"}},
			Reason:  "flat list failed (405); read parent-scoped instead"},
	}
	out := RenderImports(rs, "1.0.0")
	assert.Contains(t, out, "\n\n# Read with caveats:\n")
}

// TestRenderImportsCaveatsHeaderHasNoExtraBlankLineWhenGapsAbsent: the
// separator is only warranted when both blocks render. When "Not imported:"
// doesn't render at all, the caveats header must not gain a spurious extra
// blank line on top of the normal spacing already used elsewhere.
func TestRenderImportsCaveatsHeaderHasNoExtraBlankLineWhenGapsAbsent(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_compliance_family", Status: "ok",
			Reason: "flat list failed (405); read parent-scoped instead"},
	}
	out := RenderImports(rs, "1.0.0")
	assert.NotContains(t, out, "# Not imported:")
	assert.Contains(t, out, "}\n\n# Read with caveats:\n")
	assert.NotContains(t, out, "\n\n\n# Read with caveats:\n")
}

// --- C1: Labeler.Allocate deliberately returns the same label for two
// records sharing a (tfType, id) pair, but RenderImports must not then emit
// two import blocks with the same "to =" address -- Terraform rejects the
// whole file on a duplicate import address, so one duplicate destroys every
// other block. RenderImports must dedup on (tfType, id) and report the
// dropped count as a caveat.

func TestRenderImportsDedupsRecordsSharingAnID(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_ou", Status: "ok", Records: []Record{
			{ID: "7", Name: "Engineering"}, {ID: "7", Name: "Engineering (dup)"},
		}},
	}
	out := RenderImports(rs, "1.0.0")
	assert.Equal(t, 1, strings.Count(out, "import {"), "one duplicate id must produce exactly one block")
	assert.Contains(t, out, `to = kion_ou.engineering`)
	assert.Contains(t, out, "# Read with caveats:")
	assert.Contains(t, out, "1 duplicate record(s) skipped")
}

// --- I2: Go's %q leaves "${" and "%{" literal, but HCL reads those as
// interpolation/directive markers -- an id containing either would make the
// generated id = "..." string get reinterpreted by HCL instead of read as a
// literal value.

func TestRenderImportsEscapesHCLTemplateMarkersInID(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_iam_policy", Status: "ok", Records: []Record{
			{ID: `arn:aws:iam::123:policy/${bad}`, Name: "p"},
		}},
	}
	out := RenderImports(rs, "1.0.0")
	assert.Contains(t, out, `id = "arn:aws:iam::123:policy/$${bad}"`)
	assert.NotContains(t, out, `id = "arn:aws:iam::123:policy/${bad}"`)
}

func TestRenderImportsKeepsBothBlocksForDistinctIDs(t *testing.T) {
	t.Parallel()
	rs := []Result{
		{TFType: "kion_ou", Status: "ok", Records: []Record{
			{ID: "7", Name: "Engineering"}, {ID: "8", Name: "Sales"},
		}},
	}
	out := RenderImports(rs, "1.0.0")
	assert.Equal(t, 2, strings.Count(out, "import {"))
	assert.NotContains(t, out, "# Read with caveats:")
}
