package kionclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// inflateIntArray -
func inflateIntArray(arr []int) []interface{} {
	final := make([]interface{}, 0)

	for _, item := range arr {
		it := make(map[string]interface{})
		it["id"] = item
		final = append(final, it)
	}

	return final
}

func TestMatch(t *testing.T) {
	data := make(map[string]interface{})
	data["id"] = 200
	data["name"] = "SystemReadOnlyAccess"
	data["description"] = "The description."
	data["aws_iam_path"] = ""
	data["policy"] = "{}"
	data["owner_users"] = inflateIntArray([]int{300, 100})
	data["owner_user_groups"] = inflateIntArray([]int{0})

	////////////////////////////////////////////////////////////////////////////

	// Pass
	f := Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"100"},
	}
	filterable := Filterable{
		arr: []Filter{f},
	}
	v, err := filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	// Pass
	f = Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"300"},
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail
	f = Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"200"},
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.False(t, v)

	// Pass
	f = Filter{
		key:    "id",
		keys:   []string{"id"},
		values: []interface{}{"200"},
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	////////////////////////////////////////////////////////////////////////////

	// Pass
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`Access$`}, // Match end of word.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`Acces$`}, // Match end of word.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.False(t, v)

	// Pass
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`^System`}, // Match beginning of word.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`^ystem`}, // Match beginning of word.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.False(t, v)

	// Pass
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`Read`}, // Match any string containing text.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail - case sensitive
	f = Filter{
		key:    "name",
		keys:   []string{"name"},
		values: []interface{}{`read`}, // Match any string containing text.
		regex:  true,
	}
	filterable = Filterable{
		arr: []Filter{f},
	}
	v, err = filterable.Match(data)
	assert.NoError(t, err)
	assert.False(t, v)
}

func TestExtractValue(t *testing.T) {
	m1 := make(map[string]interface{})
	m1["id"] = 100

	m11 := make(map[string]interface{})
	m11["id"] = 300

	m2 := make([]interface{}, 0)
	m2 = append(m2, m11)
	m2 = append(m2, m1)

	m3 := make(map[string]interface{})
	m3["owner_users"] = m2
	m3["id"] = 200

	////////////////////////////////////////////////////////////////////////////

	// Pass
	f := Filter{
		key:    "id",
		keys:   []string{"id"},
		values: []interface{}{"200"},
	}
	v, err := f.DeepMatch(f.keys, m3, f.values[0])
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail
	f = Filter{
		key:    "id",
		keys:   []string{"id"},
		values: []interface{}{"1"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NoError(t, err)
	assert.False(t, v)

	////////////////////////////////////////////////////////////////////////////

	// Pass
	f = Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"100"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NoError(t, err)
	assert.True(t, v)

	// Pass
	f = Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"300"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NoError(t, err)
	assert.True(t, v)

	// Fail
	f = Filter{
		key:    "owner_users.id",
		keys:   []string{"owner_users", "id"},
		values: []interface{}{"1"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NoError(t, err)
	assert.False(t, v)

	////////////////////////////////////////////////////////////////////////////

	// Error
	f = Filter{
		key:    "id2",
		keys:   []string{"id2"},
		values: []interface{}{"1"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NotNil(t, err)
	assert.False(t, v)

	// Error
	// TODO: Should throw an error because this will never be possible.
	f = Filter{
		key:    "owner_users",
		keys:   []string{"owner_users"},
		values: []interface{}{"1"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NotNil(t, err)
	assert.False(t, v)

	// Error
	f = Filter{
		key:    "owner_users.id2",
		keys:   []string{"owner_users", "id2"},
		values: []interface{}{"1"},
	}
	v, err = f.DeepMatch(f.keys, m3, f.values[0])
	assert.NotNil(t, err)
	assert.False(t, v)
}
