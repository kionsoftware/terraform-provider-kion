package crud

import "testing"

// TestFieldPolicyWithholdsCredentialStrings pins the half of the guarantee that
// does not depend on anyone maintaining a file. A list data source exposes every
// renderable response scalar by default, so a credential the API starts
// returning would otherwise reach the schema the moment it appeared.
func TestFieldPolicyWithholdsCredentialStrings(t *testing.T) {
	t.Parallel()
	var empty FieldPolicy // no entries: the name heuristic alone must decide

	for _, field := range []string{
		"key_secret", "secret", "client_secret", "password", "passwd",
		"private_key", "aws_credential", "token", "refresh_token", "api_key", "key",
	} {
		if !empty.Hidden("account_cache", field, "types.String") {
			t.Errorf("%q is credential-shaped and must not be exposed by default", field)
		}
	}

	// A credential rides in a string. Denying anything else costs real filter
	// surface: password_needs_update is a rotation flag, not a password.
	for _, tc := range []struct{ field, modelType string }{
		{"password_needs_update", "types.Bool"},
		{"token_count", "types.Int64"},
		{"name", "types.String"},
		{"account_number", "types.String"},
		{"key_id", "types.String"}, // no pattern catches this: the file must
	} {
		if empty.Hidden("account_cache", tc.field, tc.modelType) {
			t.Errorf("%q (%s) is not a credential and must stay exposed", tc.field, tc.modelType)
		}
	}
}

// TestFieldPolicyWithholdsListedFields covers the other half: key_id looks
// ordinary, so only codegen/unexposed_fields.yaml keeps it out.
func TestFieldPolicyWithholdsListedFields(t *testing.T) {
	t.Parallel()
	p := FieldPolicy{hidden: map[string]bool{"account_cache.key_id": true}}

	if !p.Hidden("account_cache", "key_id", "types.String") {
		t.Error("a listed field must be withheld")
	}
	if p.Hidden("label", "key_id", "types.String") {
		t.Error("entries are scoped to one package")
	}
}
