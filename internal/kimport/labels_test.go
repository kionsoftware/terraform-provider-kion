package kimport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalize(t *testing.T) {
	t.Parallel()
	cases := []struct{ in, want string }{
		{"Central Services", "central_services"},
		{"AWS (GovCloud) / prod", "aws_govcloud_prod"},
		{"  spaced  out  ", "spaced_out"},
		{"2024 Budget", "_2024_budget"}, // HCL identifiers cannot start with a digit
		{"", "unnamed"},
		{"!!!", "unnamed"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, Normalize(c.in), c.in)
	}
}

func TestAllocateUsesTheName(t *testing.T) {
	t.Parallel()
	var l Labeler
	assert.Equal(t, "central_services", l.Allocate("kion_ou", "Central Services", "5"))
}

func TestAllocateSuffixesCollisions(t *testing.T) {
	t.Parallel()
	var l Labeler
	assert.Equal(t, "engineers", l.Allocate("kion_ou", "Engineers", "1"))
	assert.Equal(t, "engineers_2", l.Allocate("kion_ou", "Engineers", "2"))
	assert.Equal(t, "engineers_3", l.Allocate("kion_ou", "Engineers", "3"))
}

func TestAllocateScopesCollisionsPerType(t *testing.T) {
	t.Parallel()
	var l Labeler
	assert.Equal(t, "shared", l.Allocate("kion_ou", "Shared", "1"))
	assert.Equal(t, "shared", l.Allocate("kion_project", "Shared", "1"))
}

func TestAllocateFallsBackToID(t *testing.T) {
	t.Parallel()
	var l Labeler
	assert.Equal(t, "kion_label_9", l.Allocate("kion_label", "", "9"))
}

func TestAllocateIsIdempotent(t *testing.T) {
	t.Parallel()
	var l Labeler
	first := l.Allocate("kion_ou", "Central Services", "5")
	assert.Equal(t, first, l.Allocate("kion_ou", "Central Services", "5"))
}

func TestAllocateSanitizesAnIDWithASlash(t *testing.T) {
	t.Parallel()
	var l Labeler
	assert.Equal(t, "kion_ou_permission_mapping_3_2",
		l.Allocate("kion_ou_permission_mapping", "", "3/2"))
}
