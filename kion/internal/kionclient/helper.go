package kionclient

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FlattenStringPointer -
func FlattenStringPointer(d *schema.ResourceData, key string) *string {
	if i, ok := d.GetOk(key); ok {
		v := i.(string)
		return &v
	}
	return nil
}

// FlattenStringArray -
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

// FlattenStringArrayPointer -
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

func FilterStringArray(items []string) []string {
	arr := make([]string, 0)
	for _, item := range items {
		// Added this because compliance_check has an array with an empty value in: regions.
		if item != "" {
			arr = append(arr, item)
		}
	}

	return arr
}

// FlattenIntPointer -
func FlattenIntPointer(d *schema.ResourceData, key string) *int {
	if i, ok := d.GetOk(key); ok {
		v := i.(int)
		return &v
	}
	return nil
}

// FlattenIntArrayPointer -
func FlattenIntArrayPointer(items []interface{}) *[]int {
	arr := make([]int, 0)
	for _, item := range items {
		arr = append(arr, item.(int))
	}
	return &arr
}

// FlattenBoolArray -
func FlattenBoolArray(items []interface{}) []bool {
	arr := make([]bool, 0)
	for _, item := range items {
		arr = append(arr, item.(bool))
	}
	return arr
}

// FlattenBoolPointer -
func FlattenBoolPointer(d *schema.ResourceData, key string) *bool {
	if i, ok := d.GetOk(key); ok {
		v := i.(bool)
		return &v
	}
	return nil
}

// FlattenGenericIDArray -
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

// FlattenGenericIDPointer retrieves and converts the value associated with the given key from the schema.ResourceData.
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

// FlattenTags -
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

// FlattenAssociateLabels -
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

// InflateObjectWithID -
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

// InflateSingleObjectWithID -
func InflateSingleObjectWithID(single *ObjectWithID) interface{} {
	if single != nil {
		return single.ID
	}
	return nil
}

// InflateArrayOfIDs -
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

// InflateTags -
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

// FieldsChanged -
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

// OptionalBool -
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

// OptionalInt -
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

func OptionalString(d *schema.ResourceData, fieldname string) *string {
	v, ok := d.GetOkExists(fieldname)
	if !ok {
		return nil
	}

	ret, ok := v.(string)
	if !ok {
		return nil
	}

	return &ret
}

// AssociationChanged returns arrays of which values to change.
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

// AssociationChangedInt returns an int of a value to change.
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

// DetermineAssociations -
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

func makeMapFromArray(arr []int) map[int]bool {
	m := make(map[int]bool)
	for _, v := range arr {
		m[v] = true
	}
	return m
}

// GenerateAccTestChecksForResourceOwners -
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

// GenerateOwnerClausesForResourceTest -
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

// TestAccOUGenerateDataSourceDeclarationFilter -
func TestAccOUGenerateDataSourceDeclarationFilter(dataSourceName, localName, name string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {
			filter {
				name = "name"
				values = ["%v"]
			}
		}`, dataSourceName, localName, name)
}

// TestAccOUGenerateDataSourceDeclarationAll -
func TestAccOUGenerateDataSourceDeclarationAll(dataSourceName, localName string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {}`, dataSourceName, localName)
}

// PrintHCLConfig -
func PrintHCLConfig(config string) {
	fmt.Println("Generated HCL configuration:")
	fmt.Println(config)
}

// ConvertInterfaceSliceToIntSlice converts a slice of interfaces to a slice of integers.
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
			return nil, fmt.Errorf("unsupported type in input slice: %T", v)
		}
	}
	return output, nil
}

// GetPreviousUserAndGroupIds retrieves the previous state of user and user group IDs
// from the Terraform resource data.
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

// FindDifferences finds the differences between two slices, returning the
// elements that are present in slice1 but not in slice2.
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
type ResourceDiffSetter struct {
	Diff *schema.ResourceDiff
}

// Set wraps the SetNew method of *schema.ResourceDiff to implement the SafeSetter interface.
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

// HandleError -
func HandleError(err error) diag.Diagnostics {
	if err != nil {
		return diag.Errorf("error occurred: %v", err)
	}
	return nil
}

// AtLeastOneFieldPresent -
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

// ValidateAppRoleID is a helper function to be used in CustomizeDiff
func ValidateAppRoleID(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	// Check if app_role_id is set to 1
	if appRoleID := d.Get("app_role_id").(int); appRoleID == 1 {
		return fmt.Errorf("changing the App Role 1 via this resource is not permitted")
	}
	return nil
}
