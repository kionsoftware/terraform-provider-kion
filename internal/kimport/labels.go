package kimport

// This file allocates a unique, stable HCL label for every record
// enumeration produces.

import (
	"regexp"
	"strconv"
	"strings"
)

var nonIdent = regexp.MustCompile(`[^a-z0-9_]+`)
var repeatedUnderscore = regexp.MustCompile(`_+`)

// Normalize turns a Kion name into a valid lowercase HCL identifier.
func Normalize(s string) string {
	slug := nonIdent.ReplaceAllString(strings.ToLower(strings.TrimSpace(s)), "_")
	slug = repeatedUnderscore.ReplaceAllString(slug, "_")
	slug = strings.Trim(slug, "_")
	if slug == "" {
		return "unnamed"
	}
	if slug[0] >= '0' && slug[0] <= '9' {
		return "_" + slug
	}
	return slug
}

// Labeler hands out unique per-type labels. Uniqueness must be scoped per type
// and resolved deterministically: two OUs can each own a cloud access role named
// "Engineers", and re-running the enumerator against an unchanged install has to
// produce the same addresses or every diff is noise.
type Labeler struct {
	assigned map[string]string   // tfType + "\x00" + id -> label
	taken    map[string]struct{} // tfType + "\x00" + label
}

// Allocate returns a stable label for (tfType, id), preferring name. Identity
// is keyed on (tfType, id) alone: two records that share a (tfType, id) pair
// -- e.g. two records both with an empty id -- are treated as the same
// record and receive the same label, since id is the documented identity.
func (l *Labeler) Allocate(tfType, name, id string) string {
	if l.assigned == nil {
		l.assigned = map[string]string{}
		l.taken = map[string]struct{}{}
	}
	key := tfType + "\x00" + id
	if existing, ok := l.assigned[key]; ok {
		return existing
	}

	base := Normalize(name)
	if name == "" {
		base = Normalize(tfType + "_" + id)
	}

	label := base
	for n := 2; ; n++ {
		if _, clash := l.taken[tfType+"\x00"+label]; !clash {
			break
		}
		label = base + "_" + strconv.Itoa(n)
	}
	l.taken[tfType+"\x00"+label] = struct{}{}
	l.assigned[key] = label
	return label
}
