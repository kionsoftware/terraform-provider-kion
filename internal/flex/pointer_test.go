package flex_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"

	"terraform-provider-kion/internal/flex"
)

// Null and unknown must both be nil so omitempty leaves the field out of a raw
// body; a known zero value must still be sent.
func TestPointerFromFramework(t *testing.T) {
	t.Parallel()

	assert.Nil(t, flex.StringPointerFromFramework(types.StringNull()))
	assert.Nil(t, flex.StringPointerFromFramework(types.StringUnknown()))
	if p := flex.StringPointerFromFramework(types.StringValue("")); assert.NotNil(t, p) {
		assert.Empty(t, *p)
	}

	assert.Nil(t, flex.BoolPointerFromFramework(types.BoolNull()))
	assert.Nil(t, flex.BoolPointerFromFramework(types.BoolUnknown()))
	if p := flex.BoolPointerFromFramework(types.BoolValue(false)); assert.NotNil(t, p) {
		assert.False(t, *p)
	}

	assert.Nil(t, flex.Int64PointerFromFramework(types.Int64Null()))
	assert.Nil(t, flex.Int64PointerFromFramework(types.Int64Unknown()))
	if p := flex.Int64PointerFromFramework(types.Int64Value(0)); assert.NotNil(t, p) {
		assert.Zero(t, *p)
	}
}
