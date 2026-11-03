package crud

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// FieldPolicyPath is the hand-maintained list of response fields the data
// source must NOT expose, relative to the project root.
const FieldPolicyPath = "codegen/unexposed_fields.yaml"

// FieldPolicy decides which response-only scalars a list data source exposes.
//
// A list data source used to expose the intersection of the create request
// body and the read response, because buildListDSData walked the resource's
// attributes and looked each up in the response. Fields the server owns and a
// client cannot set were therefore unreachable from a configuration: no
// `filter` block could match on them and they were absent from the `list`
// objects. Provenance flags are the fields that suffer most -- ct_managed,
// system_managed_policy and their siblings are exactly what a create body has
// no reason to carry, and exactly what tells a shipped record from a
// customer's.
//
// The default is now to expose every renderable response scalar, and this
// policy is the exception list. Inverting it that way matters: the previous
// behavior hid new fields silently, whereas hiding one now takes an entry in
// a reviewed file.
//
// Two things withhold a field, because they fail differently. The file names
// exactly what to hide and covers a credential no pattern would recognize
// (account_cache.key_id); secretName recognizes a credential-shaped string
// nobody listed, which is what protects a secret the API has not returned yet.
type FieldPolicy struct {
	hidden map[string]bool
}

// LoadFieldPolicy reads FieldPolicyPath. A missing file is an error rather
// than an empty policy: silently exposing credentials because a file went
// missing is the one failure mode worth being loud about.
func LoadFieldPolicy(root string) (FieldPolicy, error) {
	path := filepath.Join(root, FieldPolicyPath)
	raw, err := os.ReadFile(path)
	if err != nil {
		return FieldPolicy{}, fmt.Errorf("reading %s: %w", FieldPolicyPath, err)
	}
	var entries map[string]string
	if err := yaml.Unmarshal(raw, &entries); err != nil {
		return FieldPolicy{}, fmt.Errorf("parsing %s: %w", FieldPolicyPath, err)
	}
	hidden := make(map[string]bool, len(entries))
	for key := range entries {
		hidden[key] = true
	}
	return FieldPolicy{hidden: hidden}, nil
}

// Hidden reports whether pkg's field is withheld from the data source.
// modelType is the framework type the field would project to; a credential can
// only ride in a string, so the name heuristic applies to nothing else.
func (p FieldPolicy) Hidden(pkg, field, modelType string) bool {
	if p.hidden[pkg+"."+field] {
		return true
	}
	return modelType == "types.String" && secretName(field)
}

// secretNameParts are the substrings that mark a field name as credential-
// shaped. They are matched against response-only fields, never against a
// resource attribute, so nothing a configuration already sets is affected.
var secretNameParts = []string{
	"secret", "password", "passwd", "private_key", "credential", "token",
}

// secretName reports whether a string field's name looks like a credential.
//
// This is the structural half of the guarantee that account_cache's key_secret
// stays unexposed: exposure is default-deny for a credential-shaped string, so
// a newly returned secret is withheld even if nobody thought to list it.
// unexposed_fields.yaml still carries key_id, which no pattern can recognize --
// a list and a pattern cover different failure modes, and neither is enough
// alone.
func secretName(field string) bool {
	f := strings.ToLower(field)
	for _, part := range secretNameParts {
		if strings.Contains(f, part) {
			return true
		}
	}
	return f == "key" || strings.HasSuffix(f, "_key")
}
