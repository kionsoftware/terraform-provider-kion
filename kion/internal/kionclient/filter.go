package kionclient

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Filterable -
type Filterable struct {
	arr []Filter
}

// NewFilterable -
func NewFilterable(d *schema.ResourceData) *Filterable {
	arr := make([]Filter, 0)

	v, oclient := d.GetOk("filter")
	if !ok {
		return nil
	}

	filterList := v.([]interface{})

	for _, v := range filterList {
		fi := v.(map[string]interface{})

		filterName := fi["name"].(string)
		filterValues := fi["values"].([]interface{})
		filterRegex := fi["regex"].(bool)

		f := Filter{
			key:    filterName,
			keys:   strings.Split(filterName, "."),
			values: filterValues,
			regex:  filterRegex,
		}

		arr = append(arr, f)
	}

	return &Filterable{
		arr: arr,
	}
}

// Match -
func (f *Filterable) Match(m map[string]interface{}) (bool, error) {
	// Match if filterable is because there is no filter.
	if f == nil {
		return true, nil
	}

	// Loop through each filter.
	for _, filter := range f.arr {
		found := false
		for _, filterValue := range filter.values {
			match, err := filter.DeepMatch(filter.keys, m, filterValue)
			if err != nil {
				return false, err
			} else if match {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}

	return true, nil
}

// Filter -
type Filter struct {
	key  string
	keys []string
	// These will always be an array of strings so when doing a comparison,
	// you have to convert to a string using: fmt.Sprint().
	values []interface{}
	regex  bool
}

// DeepMatch -
func (f *Filter) DeepMatch(keys []string, m map[string]interface{}, filterValue interface{}) (bool, error) {
	val, oclient := m[keys[0]]
	if !ok {
		return false, errors.New("filter is not found: " + keys[0] + fmt.Sprintf(" | %#v", m))
	}

	if len(keys) == 1 {
		// Catch a user error if the filter is comparing against an array
		// ex. Using a filter of 'owner_users' instead of 'owner_users.id'
		if _, oclient := val.([]interface{}); ok {
			return false, fmt.Errorf("filter key (%v) references an array instead of a field: %v", f.key, fmt.Sprint(val))
		}
		// If set as a regex, then compare against it.
		if f.regex {
			re, err := regexp.Compile(fmt.Sprint(filterValue))
			if err != nil {
				return false, fmt.Errorf("invalid regular expression '%v' for '%v' filter", filterValue, f.key)
			}
			return re.MatchString(fmt.Sprint(val)), nil
		}
		// filterValue will always be a string so compare accordingly.
		return fmt.Sprint(val) == filterValue, nil
	}

	if x, oclient := val.([]interface{}); ok {
		// If the field is an array, then determine if one of the values matches.
		for _, i := range x {
			vmap := i.(map[string]interface{})

			match, err := f.DeepMatch(keys[1:], vmap, filterValue)
			if err != nil {
				return false, err
			} else if match {
				return true, nil
			}
		}
	}

	return false, nil
}

func isZero(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}
