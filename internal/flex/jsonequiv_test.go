package flex_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"terraform-provider-kion/internal/flex"
)

func TestJSONEquivalent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a, b string
		want bool
	}{
		// The case this exists for: Kion re-marshals an Azure role's permissions
		// and fills in the two arrays the practitioner omitted.
		{
			name: "api adds empty arrays",
			a:    `{"actions":["Microsoft.Resources/subscriptions/read"],"notActions":[]}`,
			b:    `{"actions":["Microsoft.Resources/subscriptions/read"],"dataActions":[],"notActions":[],"notDataActions":[]}`,
			want: true,
		},
		{name: "identical", a: `{"a":[1]}`, b: `{"a":[1]}`, want: true},
		{name: "added empty object", a: `{"a":1}`, b: `{"a":1,"b":{}}`, want: true},
		{name: "added null", a: `{"a":1}`, b: `{"a":1,"b":null}`, want: true},
		{name: "key order is irrelevant", a: `{"a":1,"b":2}`, b: `{"b":2,"a":1}`, want: true},

		// Anything that is a real difference must stay one, or this masks drift.
		{name: "added non-empty key", a: `{"a":1}`, b: `{"a":1,"b":[2]}`, want: false},
		{name: "differing value", a: `{"a":1}`, b: `{"a":2}`, want: false},
		{name: "differing array contents", a: `{"a":["x"]}`, b: `{"a":["y"]}`, want: false},
		{name: "array order", a: `{"a":["x","y"]}`, b: `{"a":["y","x"]}`, want: false},
		{name: "array length", a: `{"a":["x"]}`, b: `{"a":["x","y"]}`, want: false},
		{name: "removed non-empty key", a: `{"a":1,"b":[2]}`, b: `{"a":1}`, want: false},
		{name: "nested difference", a: `{"a":{"b":1}}`, b: `{"a":{"b":2}}`, want: false},
		{name: "nested empty addition", a: `{"a":{"b":1}}`, b: `{"a":{"b":1,"c":[]}}`, want: true},

		// Non-JSON is compared verbatim rather than guessed at.
		{name: "both invalid, identical", a: `not json`, b: `not json`, want: true},
		{name: "both invalid, different", a: `not json`, b: `also not json`, want: false},
		{name: "one invalid", a: `{"a":1}`, b: `not json`, want: false},
		{name: "empty strings", a: ``, b: ``, want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, flex.JSONEquivalent(tc.a, tc.b))
			// The relation must not depend on argument order.
			assert.Equal(t, tc.want, flex.JSONEquivalent(tc.b, tc.a), "not symmetric")
		})
	}
}
