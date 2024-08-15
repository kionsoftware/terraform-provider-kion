package kionclient

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Filterable holds an array of filters that can be applied to data.
type Filterable struct {
	arr []Filter
}

func NewFilterable(d *schema.ResourceData) *Filterable {
	arr := make([]Filter, 0)

	filterList := d.Get("filter").([]interface{})
	for _, v := range filterList {
		fi := v.(map[string]interface{})

		filterName, nameOk := fi["name"].(string)
		filterValues, valuesOk := fi["values"].([]interface{})
		filterRegex, regexOk := fi["regex"].(bool)

		if nameOk && valuesOk {
			f := Filter{
				key:    filterName,
				keys:   strings.Split(filterName, "."),
				values: filterValues,
				regex:  regexOk && filterRegex,
			}
			arr = append(arr, f)
		} else {
			return nil
		}
	}

	return &Filterable{
		arr: arr,
	}
}

// Match applies the filters to the provided map of data. It returns true if the data matches all filters, otherwise false.
// If no filters are present, it matches everything by default.
func (f *Filterable) Match(m map[string]interface{}) (bool, error) {
	if f == nil || len(f.arr) == 0 {
		return true, nil
	}

	for _, filter := range f.arr {
		match := false
		for _, filterValue := range filter.values {
			matched, err := filter.DeepMatch(filter.keys, m, filterValue)
			if err != nil {
				return false, err
			}
			if matched {
				match = true
				break
			}
		}
		if !match {
			return false, nil
		}
	}

	return true, nil
}

// Filter represents a single filter criterion that can be applied to data.
type Filter struct {
	key    string
	keys   []string
	values []interface{}
	regex  bool
}

// DeepMatch is a recursive function used to match deeply nested fields within a map.
// It supports both exact matching and regex-based matching.
func (f *Filter) DeepMatch(keys []string, m map[string]interface{}, filterValue interface{}) (bool, error) {
	val, ok := m[keys[0]]
	if !ok {
		return false, errors.New("filter not found: " + keys[0] + fmt.Sprintf(" | %#v", m))
	}

	if len(keys) == 1 {
		if _, ok := val.([]interface{}); ok {
			return false, fmt.Errorf("filter key (%v) references an array instead of a field: %v", f.key, fmt.Sprint(val))
		}
		if f.regex {
			re, err := regexp.Compile(fmt.Sprint(filterValue))
			if err != nil {
				return false, fmt.Errorf("invalid regular expression '%v' for '%v' filter", filterValue, f.key)
			}
			return re.MatchString(fmt.Sprint(val)), nil
		}
		return fmt.Sprint(val) == fmt.Sprint(filterValue), nil
	}

	if x, ok := val.([]interface{}); ok {
		for _, i := range x {
			vmap, ok := i.(map[string]interface{})
			if !ok {
				continue
			}

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
