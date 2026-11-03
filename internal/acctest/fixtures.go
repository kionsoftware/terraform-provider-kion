package acctest

import "fmt"

// Shared prerequisite configuration.
//
// A resource whose test needs a funding source has to build one, and a funding
// source is the most demanding prerequisite in the provider: it requires an OU,
// and two permission schemes of different types. Four packages each wrote their
// own version, and all four were wrong in the same two ways -- no ou_id, and a
// hardcoded permission_scheme_id that only existed on one seeded install:
//
//	Internal Server Error: ou id required
//	Bad Request: app policy type not valid for this object
//
// Keeping one copy here means the next resource that needs a funding source
// inherits a correct one, and a change in what Kion requires is made once.

// FundingSourceConfig returns HCL declaring kion_funding_source.test_fs, plus
// the OU and permission schemes it depends on, all named from prefix.
//
// The two schemes are not interchangeable: kion_ou takes an "ou"-typed scheme
// and kion_funding_source a "funding_source"-typed one, and Kion rejects the
// pairing the other way round. Reference the funding source as
// kion_funding_source.test_fs.id.
func FundingSourceConfig(prefix string) string {
	return fmt.Sprintf(`
resource "kion_permission_scheme" "test_fs_ou_perm" {
  name = "%[1]s-ou-perm"
  type = "ou"
}

resource "kion_permission_scheme" "test_fs_perm" {
  name = "%[1]s-fs-perm"
  type = "funding_source"
}

resource "kion_ou" "test_fs_ou" {
  name                 = "%[1]s-ou"
  parent_ou_id         = 0
  permission_scheme_id = kion_permission_scheme.test_fs_ou_perm.id
  owner_user_ids       = [1]
}

resource "kion_funding_source" "test_fs" {
  name                 = "%[1]s-fs"
  amount               = 1000.00
  start_datecode       = "2026-01"
  end_datecode         = "2026-12"
  owner_user_ids       = [1]
  ou_id                = kion_ou.test_fs_ou.id
  permission_scheme_id = kion_permission_scheme.test_fs_perm.id
}
`, prefix)
}

// ShortName truncates a random test name to fit a column narrower than the
// 17 characters RandomWithPrefix produces.
func ShortName(rName string, maxLen int) string {
	if len(rName) > maxLen {
		return rName[:maxLen]
	}
	return rName
}
