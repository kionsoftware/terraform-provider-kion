package kionclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FlattenStringPointer retrieves a string value from the schema.ResourceData by its key
// and returns a pointer to that string. If the key does not exist, it returns nil.
func FlattenStringPointer(d *schema.ResourceData, key string) *string {
	if i, ok := d.GetOk(key); ok {
		v := i.(string)
		return &v
	}
	return nil
}

// FlattenStringArray converts an array of interfaces to an array of strings,
// filtering out any empty string values.
func FlattenStringArray(items []interface{}) []string {
	arr := make([]string, 0)
	for _, item := range items {
		v := item.(string)
		// Filter out empty values
		if v != "" {
			arr = append(arr, v)
		}
	}
	return arr
}

// FlattenStringArrayPointer retrieves an array of strings from schema.ResourceData
// by its key, filters out empty values, and returns a pointer to the resulting slice.
// If the key does not exist, it returns nil.
func FlattenStringArrayPointer(d *schema.ResourceData, key string) *[]string {
	if i, ok := d.GetOk(key); ok {
		v := i.([]string)
		arr := make([]string, 0)
		for _, item := range v {
			// Filter out empty values
			if item != "" {
				arr = append(arr, item)
			}
		}
		return &arr
	}
	return nil
}

// FilterStringArray filters out empty strings from an array of strings
// and returns a new array containing only non-empty strings.
func FilterStringArray(items []string) []string {
	arr := make([]string, 0)
	for _, item := range items {
		// Filter out empty values
		if item != "" {
			arr = append(arr, item)
		}
	}
	return arr
}

// FlattenIntPointer retrieves an integer value from schema.ResourceData by its key
// and returns a pointer to that integer. If the key does not exist, it returns nil.
func FlattenIntPointer(d *schema.ResourceData, key string) *int {
	if i, ok := d.GetOk(key); ok {
		v := i.(int)
		return &v
	}
	return nil
}

// FlattenIntArrayPointer converts an array of interfaces to an array of integers,
// and returns a pointer to the resulting slice.
func FlattenIntArrayPointer(items []interface{}) *[]int {
	arr := make([]int, 0)
	for _, item := range items {
		arr = append(arr, item.(int))
	}
	return &arr
}

// FlattenBoolArray converts an array of interfaces to an array of booleans.
func FlattenBoolArray(items []interface{}) []bool {
	arr := make([]bool, 0)
	for _, item := range items {
		arr = append(arr, item.(bool))
	}
	return arr
}

// FlattenBoolPointer retrieves a boolean value from schema.ResourceData by its key
// and returns a pointer to that boolean. If the key does not exist, it returns nil.
func FlattenBoolPointer(d *schema.ResourceData, key string) *bool {
	if i, ok := d.GetOk(key); ok {
		v := i.(bool)
		return &v
	}
	return nil
}

// FlattenGenericIDArray retrieves a list of objects from schema.ResourceData by its key,
// extracts the "id" field from each object, and returns a slice of these IDs.
func FlattenGenericIDArray(d *schema.ResourceData, key string) []int {
	uid := d.Get(key).([]interface{})
	uids := make([]int, 0)
	for _, item := range uid {
		v, ok := item.(map[string]interface{})
		if ok {
			uids = append(uids, v["id"].(int))
		}
	}
	return uids
}

// FlattenGenericIDPointer retrieves a list of objects or a schema.Set from schema.ResourceData by its key,
// extracts the "id" field from each object, and returns a pointer to a slice of these IDs.
func FlattenGenericIDPointer(d *schema.ResourceData, key string) *[]int {
	uid := d.Get(key)

	switch v := uid.(type) {
	case []interface{}:
		uids := make([]int, len(v))
		for i, item := range v {
			uids[i] = item.(int)
		}
		return &uids
	case *schema.Set:
		setList := v.List()
		uids := make([]int, len(setList))
		for i, item := range setList {
			m := item.(map[string]interface{})
			uids[i] = m["id"].(int)
		}
		return &uids
	default:
		return nil
	}
}

// FlattenTags retrieves a map of tags from schema.ResourceData by its key and
// returns a pointer to a slice of Tag objects, each containing a key and value.
func FlattenTags(d *schema.ResourceData, key string) *[]Tag {
	tagMap := d.Get(key).(map[string]interface{})
	tags := make([]Tag, 0)
	for k, v := range tagMap {
		tags = append(tags, Tag{
			Key:   k,
			Value: v.(string),
		})
	}
	return &tags
}

// FlattenAssociateLabels retrieves a map of associate labels from schema.ResourceData by its key
// and returns a pointer to a slice of AssociateLabel objects.
func FlattenAssociateLabels(d *schema.ResourceData, key string) *[]AssociateLabel {
	labelMap := d.Get(key).(map[string]interface{})
	labels := make([]AssociateLabel, len(labelMap))
	var idx int
	for k, v := range labelMap {
		labels[idx].Key = k
		labels[idx].Value = v.(string)
		idx++
	}
	return &labels
}

// InflateObjectWithID converts a slice of ObjectWithID to a slice of interfaces
// where each object is represented as a map with an "id" field.
func InflateObjectWithID(arr []ObjectWithID) []interface{} {
	if arr == nil {
		return make([]interface{}, 0)
	}

	final := make([]interface{}, 0)
	for _, item := range arr {
		it := make(map[string]interface{})
		it["id"] = item.ID
		final = append(final, it)
	}
	return final
}

// InflateSingleObjectWithID converts an ObjectWithID to an interface containing its "id" field.
func InflateSingleObjectWithID(single *ObjectWithID) interface{} {
	if single != nil {
		return single.ID
	}
	return nil
}

// InflateArrayOfIDs converts a slice of integers to a slice of interfaces
// where each integer is represented as a map with an "id" field.
func InflateArrayOfIDs(arr []int) []interface{} {
	if arr == nil {
		return make([]interface{}, 0)
	}

	final := make([]interface{}, 0)
	for _, item := range arr {
		it := make(map[string]interface{})
		it["id"] = item
		final = append(final, it)
	}
	return final
}

// InflateTags converts a slice of Tag objects to a map where the keys are the tag keys
// and the values are the tag values.
func InflateTags(arr []Tag) map[string]string {
	if arr == nil {
		return nil
	}

	final := make(map[string]string)
	for _, item := range arr {
		final[item.Key] = item.Value
	}
	return final
}

// FieldsChanged compares two maps of resource attributes and checks if any of the specified fields
// have changed between the old and new versions. It returns the new map, the first field that changed,
// and a boolean indicating whether any field changed.
func FieldsChanged(iOld, iNew interface{}, fields []string) (map[string]interface{}, string, bool) {
	mOld := iOld.(map[string]interface{})
	mNew := iNew.(map[string]interface{})

	for _, v := range fields {
		if mNew[v] != mOld[v] {
			return mNew, v, true
		}
	}
	return mNew, "", false
}

// OptionalBool retrieves a boolean value from schema.ResourceData by its field name and
// returns a pointer to that value. If the field is not set or is not a boolean, it returns nil.
func OptionalBool(d *schema.ResourceData, fieldname string) *bool {
	b, ok := d.GetOkExists(fieldname)
	if !ok {
		return nil
	}

	ret, ok := b.(bool)
	if !ok {
		return nil
	}
	return &ret
}

// OptionalInt retrieves an integer value from schema.ResourceData by its field name and
// returns a pointer to that value. If the field is not set or is not an integer, it returns nil.
func OptionalInt(d *schema.ResourceData, fieldname string) *int {
	v, ok := d.GetOkExists(fieldname)
	if !ok {
		return nil
	}

	ret, ok := v.(int)
	if !ok {
		return nil
	}
	return &ret
}

// OptionalValue retrieves a value from schema.ResourceData by its field name and returns a pointer to that value.
// The function uses type assertion to handle different types like int, bool, and string.
func OptionalValue[T any](d *schema.ResourceData, fieldname string) *T {
	v, ok := d.GetOkExists(fieldname)
	if !ok {
		return nil
	}

	// Use type assertion to check the type of the value
	if ret, ok := v.(T); ok {
		return &ret
	}

	return nil
}

// AssociationChanged compares the old and new values of a field that contains an array of IDs
// (e.g., user or group IDs) and determines which IDs were added, removed, or changed.
// It returns slices of IDs to add and remove, a boolean indicating if there was a change, and any error encountered.
func AssociationChanged(d *schema.ResourceData, fieldname string) ([]int, []int, bool, error) {
	isChanged := false
	io, in := d.GetChange(fieldname)

	_, isTypeSet := io.(*schema.Set)
	if isTypeSet {
		io = io.(*schema.Set).List()
		in = in.(*schema.Set).List()
	}

	ownerOld := io.([]interface{})
	oldIDs, err := ConvertInterfaceSliceToIntSlice(ownerOld)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to convert old IDs in field '%s': %w", fieldname, err)
	}

	ownerNew := in.([]interface{})
	newIDs, err := ConvertInterfaceSliceToIntSlice(ownerNew)
	if err != nil {
		return nil, nil, false, fmt.Errorf("failed to convert new IDs in field '%s': %w", fieldname, err)
	}

	arrUserAdd, arrUserRemove, changed := determineAssociations(newIDs, oldIDs)
	if changed {
		isChanged = true
	}

	return arrUserAdd, arrUserRemove, isChanged, nil
}

// AssociationChangedInt compares the old and new values of a field that contains a single integer ID
// and determines if the ID has changed. It returns the new and old ID, a boolean indicating if there was a change, and any error encountered.
func AssociationChangedInt(d *schema.ResourceData, fieldname string) (*int, *int, bool, error) {
	isChanged := false
	io, in := d.GetChange(fieldname)

	if in != io {
		isChanged = true
		if in == nil || in == 0 {
			old := io.(int)
			return nil, &old, isChanged, nil
		}
		newvalue := in.(int)
		return &newvalue, nil, isChanged, nil
	}
	return nil, nil, isChanged, nil
}

// determineAssociations compares two slices of integers and determines which values are present in the source slice
// but not in the destination slice (to add) and which are present in the destination slice but not in the source slice (to remove).
// It returns slices of IDs to add and remove, and a boolean indicating if there was a change.
func determineAssociations(src []int, dest []int) (arrAdd []int, arrRemove []int, isChanged bool) {
	mSrc := makeMapFromArray(src)
	mDest := makeMapFromArray(dest)

	arrAdd = make([]int, 0)
	arrRemove = make([]int, 0)
	isChanged = false

	for v := range mSrc {
		if _, found := mDest[v]; !found {
			arrAdd = append(arrAdd, v)
			isChanged = true
		}
	}

	for v := range mDest {
		if _, found := mSrc[v]; !found {
			arrRemove = append(arrRemove, v)
			isChanged = true
		}
	}
	return arrAdd, arrRemove, isChanged
}

// makeMapFromArray converts a slice of integers to a map where each integer is a key with a boolean value of true.
func makeMapFromArray(arr []int) map[int]bool {
	m := make(map[int]bool)
	for _, v := range arr {
		m[v] = true
	}
	return m
}

// GenerateAccTestChecksForResourceOwners generates a list of Terraform acceptance test check functions
// to verify that the specified resource owners (user IDs and user group IDs) are correctly set in a resource's state.
func GenerateAccTestChecksForResourceOwners(resourceType, resourceName string, ownerUserIds, ownerUserGroupIds *[]int) []resource.TestCheckFunc {
	var funcs []resource.TestCheckFunc

	if ownerUserIds != nil {
		for idx, id := range *ownerUserIds {
			funcs = append(funcs, resource.TestCheckResourceAttr(
				resourceType+"."+resourceName,
				fmt.Sprintf("owner_users.%v.id", idx),
				fmt.Sprint(id),
			))
		}
	}

	if ownerUserGroupIds != nil {
		for idx, id := range *ownerUserGroupIds {
			funcs = append(funcs, resource.TestCheckResourceAttr(
				resourceType+"."+resourceName,
				fmt.Sprintf("owner_user_groups.%v.id", idx),
				fmt.Sprint(id),
			))
		}
	}

	return funcs
}

// GenerateOwnerClausesForResourceTest generates a string containing HCL code for setting
// owner users and owner user groups in a Terraform resource test case.
func GenerateOwnerClausesForResourceTest(ownerUserIds, ownerUserGroupIds *[]int) (ownerClauses string) {
	if ownerUserIds != nil {
		for _, id := range *ownerUserIds {
			ownerClauses += fmt.Sprintf("\nowner_users { id = %v }", id)
		}
	}

	if ownerUserGroupIds != nil {
		for _, id := range *ownerUserGroupIds {
			ownerClauses += fmt.Sprintf("\nowner_user_groups { id = %v }", id)
		}
	}

	return
}

// TestAccOUGenerateDataSourceDeclarationFilter generates a string containing HCL code
// for declaring a Terraform data source with a filter based on the "name" attribute.
func TestAccOUGenerateDataSourceDeclarationFilter(dataSourceName, localName, name string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {
			filter {
				name = "name"
				values = ["%v"]
			}
		}`, dataSourceName, localName, name)
}

// TestAccOUGenerateDataSourceDeclarationAll generates a string containing HCL code
// for declaring a Terraform data source without any filters.
func TestAccOUGenerateDataSourceDeclarationAll(dataSourceName, localName string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {}`, dataSourceName, localName)
}

// PrintHCLConfig prints a provided HCL configuration string to the console.
func PrintHCLConfig(config string) {
	fmt.Println("Generated HCL configuration:")
	fmt.Println(config)
}

// ConvertInterfaceSliceToIntSlice converts a slice of interfaces to a slice of integers.
// It handles the conversion and checks if the elements are integers or maps with an "id" field.
func ConvertInterfaceSliceToIntSlice(input []interface{}) ([]int, error) {
	output := make([]int, len(input))
	for i, v := range input {
		// Check if the element is an integer
		switch val := v.(type) {
		case int:
			output[i] = val
		case map[string]interface{}:
			// Handle complex type, extract ID or handle accordingly
			if id, ok := val["id"].(int); ok {
				output[i] = id
			} else {
				return nil, fmt.Errorf("expected 'id' field to be int, got: %v", val)
			}
		default:
			// ConvertInterfaceSliceToIntSlice (continued)...
			return nil, fmt.Errorf("unsupported type in input slice: %T", v)
		}
	}
	return output, nil
}

// GetPreviousUserAndGroupIds retrieves the previous state of user and user group IDs
// from the Terraform resource data. It checks if the "user_ids" and "user_group_ids"
// fields have changed and converts the previous values to slices of integers.
func GetPreviousUserAndGroupIds(d *schema.ResourceData) ([]int, []int, error) {
	var prevUserIds, prevUserGroupIds []int
	var err error

	// Check if the "user_ids" field has changed
	if d.HasChange("user_ids") {
		// Get the previous value of the "user_ids" field
		oldValue, _ := d.GetChange("user_ids")
		// Convert the previous value to a slice of integers
		prevUserIds, err = ConvertInterfaceSliceToIntSlice(oldValue.([]interface{}))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert previous user IDs: %w", err)
		}
	}

	// Check if the "user_group_ids" field has changed
	if d.HasChange("user_group_ids") {
		// Get the previous value of the "user_group_ids" field
		oldValue, _ := d.GetChange("user_group_ids")
		// Convert the previous value to a slice of integers
		prevUserGroupIds, err = ConvertInterfaceSliceToIntSlice(oldValue.([]interface{}))
		if err != nil {
			return nil, nil, fmt.Errorf("failed to convert previous user group IDs: %w", err)
		}
	}

	return prevUserIds, prevUserGroupIds, nil
}

// FindDifferences finds the differences between two slices, returning the elements
// that are present in slice1 but not in slice2. This function is generic and works with any comparable type.
func FindDifferences[T comparable](slice1, slice2 []T) []T {
	// Create a set from the second slice for efficient lookups
	set := make(map[T]bool)
	for _, v := range slice2 {
		set[v] = true
	}

	var diff []T
	// Iterate over the first slice and find elements not present in the second slice
	for _, v := range slice1 {
		if !set[v] {
			diff = append(diff, v)
		}
	}
	return diff
}

// ResourceDiffSetter is a wrapper around *schema.ResourceDiff to implement the SafeSetter interface.
// This allows it to be used in contexts where a consistent interface for setting values is needed.
type ResourceDiffSetter struct {
	Diff *schema.ResourceDiff
}

// Set wraps the SetNew method of *schema.ResourceDiff to implement the SafeSetter interface.
// It allows for setting new values during the resource diff phase in Terraform.
func (r *ResourceDiffSetter) Set(key string, value interface{}) error {
	return r.Diff.SetNew(key, value)
}

// SafeSetter is an interface that abstracts the behavior of setting a key-value pair
// in Terraform's schema. It is implemented by both *schema.ResourceData and a custom wrapper
// around *schema.ResourceDiff, allowing for a unified handling of schema mutations across
// different Terraform lifecycle phases.
type SafeSetter interface {
	Set(key string, value interface{}) error
}

// SafeSet handles setting Terraform schema values, centralizing error reporting and ensuring non-nil values.
// It takes a SafeSetter interface (which can be either a ResourceDiffSetter or ResourceData),
// the key and value to set, and a summary message for diagnostics in case of an error.
func SafeSet(d SafeSetter, key string, value interface{}, summary string) diag.Diagnostics {
	var diags diag.Diagnostics

	if value != nil {
		if err := d.Set(key, value); err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  summary,
				Detail:   fmt.Sprintf("Error setting %s: %v", key, err),
			})
		}
	}
	return diags
}

// ParseResourceID parses a resource ID string into its components based on the expected number of parts.
// It returns the extracted ID components as integers and any error encountered during parsing.
func ParseResourceID(resourceID string, expectedParts int, idNames ...string) ([]int, error) {
	parts := strings.Split(resourceID, "-")
	if len(parts) != expectedParts {
		return nil, fmt.Errorf("invalid resource ID format, expected %s with %d parts", strings.Join(idNames, "-"), expectedParts)
	}

	ids := make([]int, len(idNames))
	for i, name := range idNames {
		id, err := strconv.Atoi(parts[i])
		if err != nil {
			return nil, fmt.Errorf("invalid %s, must be an integer", name)
		}
		ids[i] = id
	}

	return ids, nil
}

// HandleError converts an error to a diagnostic object.
// It is a simple utility function that returns a diagnostic with an error message if an error is present.
func HandleError(err error) diag.Diagnostics {
	if err != nil {
		return diag.Errorf("error occurred: %v", err)
	}
	return nil
}

// AtLeastOneFieldPresent checks a map of fields (by their names) to ensure that at least one field has a value.
// It returns an error if none of the fields contain a value, indicating that at least one must be specified.
func AtLeastOneFieldPresent(fields map[string]interface{}) error {
	for _, field := range fields {
		switch v := field.(type) {
		case []uint:
			if len(v) > 0 {
				return nil
			}
		case *schema.Set:
			if v.Len() > 0 {
				return nil
			}
		default:
			// Add other cases as needed for different field types
		}
	}

	var fieldNames []string
	for name := range fields {
		fieldNames = append(fieldNames, name)
	}

	return fmt.Errorf("at least one of the following fields must be specified: %v", fieldNames)
}

// ValidateAppRoleID is a helper function to be used in CustomizeDiff.
// It checks if the app_role_id is set to 1 and returns an error if so,
// preventing changes to the specific App Role ID 1 via this resource.
func ValidateAppRoleID(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Check if app_role_id is set to 1
	if appRoleID := d.Get("app_role_id").(int); appRoleID == 1 {
		return fmt.Errorf("changing the App Role 1 via this resource is not permitted")
	}
	return nil
}

// Custom Variable Types for Custom Variables
const (
	TypeString = "string"
	TypeList   = "list"
	TypeMap    = "map"
)

// NormalizeCvValue normalizes a custom variable value based on its type
func NormalizeCvValue(v string, cvType string) (string, error) {
	switch cvType {
	case TypeString:
		// For strings, wrap in the expected format
		return fmt.Sprintf(`{"value":%q}`, v), nil

	case TypeList, TypeMap:
		// For lists and maps, try to parse as JSON first
		var parsed interface{}
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return "", fmt.Errorf("invalid JSON for type %s: %v", cvType, err)
		}
		// Wrap in the expected format
		wrapper := map[string]interface{}{
			"value": parsed,
		}
		bytes, err := json.Marshal(wrapper)
		if err != nil {
			return "", err
		}
		return string(bytes), nil

	default:
		return "", fmt.Errorf("unsupported custom variable type: %s", cvType)
	}
}

// PackCvValueIntoJsonStr converts the API response back to the appropriate format
func PackCvValueIntoJsonStr(value interface{}, cvType string) (string, error) {
	if value == nil {
		return "", nil
	}

	switch cvType {
	case TypeString:
		switch v := value.(type) {
		case string:
			return v, nil
		default:
			return "", fmt.Errorf("expected string value, got %T", value)
		}

	case TypeList:
		switch v := value.(type) {
		case []interface{}:
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("failed to marshal list value: %v", err)
			}
			return string(bytes), nil
		case string:
			// If it's already a JSON string, validate it's a list
			var list []interface{}
			if err := json.Unmarshal([]byte(v), &list); err != nil {
				return "", fmt.Errorf("invalid list JSON: %v", err)
			}
			return v, nil
		default:
			return "", fmt.Errorf("expected list value, got %T", value)
		}

	case TypeMap:
		switch v := value.(type) {
		case map[string]interface{}:
			bytes, err := json.Marshal(v)
			if err != nil {
				return "", fmt.Errorf("failed to marshal map value: %v", err)
			}
			return string(bytes), nil
		case string:
			// If it's already a JSON string, validate it's a map
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(v), &m); err != nil {
				return "", fmt.Errorf("invalid map JSON: %v", err)
			}
			return v, nil
		default:
			return "", fmt.Errorf("expected map value, got %T", value)
		}

	default:
		return "", fmt.Errorf("unsupported custom variable type: %s", cvType)
	}
}

// UnpackCvValueJsonStr converts the input value based on the custom variable type
func UnpackCvValueJsonStr(input interface{}, cvType string) (interface{}, error) {
	switch cvType {
	case TypeString:
		if str, ok := input.(string); ok {
			// For strings, we need to send just the raw string
			return str, nil
		}
		return nil, fmt.Errorf("expected string value for type '%s', got %T", TypeString, input)

	case TypeList:
		switch v := input.(type) {
		case []interface{}:
			return v, nil
		case []string:
			// Convert []string to []interface{}
			result := make([]interface{}, len(v))
			for i, s := range v {
				result[i] = s
			}
			return result, nil
		default:
			return nil, fmt.Errorf("expected list value for type '%s', got %T", TypeList, input)
		}

	case TypeMap:
		switch v := input.(type) {
		case map[string]interface{}:
			return v, nil
		case map[string]string:
			// Convert map[string]string to map[string]interface{}
			result := make(map[string]interface{})
			for k, v := range v {
				result[k] = v
			}
			return result, nil
		default:
			return nil, fmt.Errorf("expected map value for type '%s', got %T", TypeMap, input)
		}

	default:
		return nil, fmt.Errorf("unsupported custom variable type: %s", cvType)
	}
}
