package kionclient

import (
	"fmt"

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
		// Add this because compliance_check has an array with an empty value in: regions.
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
			v := item
			// Add this because compliance_check has an array with an empty value in: regions.
			if v != "" {
				arr = append(arr, v)
			}
		}
		return &arr
	}

	return nil
}

// FilterStringArray -
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

// FlattenIntArray -
func FlattenIntArray(items []interface{}) []int {
	arr := make([]int, 0)
	for _, item := range items {
		arr = append(arr, item.(int))
	}

	return arr
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

func ConvertToIntSlice(interfaceSlice []interface{}) []int {
	intSlice := make([]int, len(interfaceSlice))
	for i, v := range interfaceSlice {
		intSlice[i] = v.(int)
	}
	return intSlice
}

// FlattenGenericIDPointer retrieves and converts the value associated with the given key from the schema.ResourceData.
// It handles different types of input ([]interface{} and *schema.Set) and returns a pointer to a slice of integers.
func FlattenGenericIDPointer(d *schema.ResourceData, key string) *[]int {
	// Retrieve the value from the resource data using the provided key
	uid := d.Get(key)

	// Determine the type of the retrieved value
	switch v := uid.(type) {
	// Handle the case where the value is a slice of interfaces
	case []interface{}:
		// Create a slice of integers with the same length as the input slice
		uids := make([]int, len(v))
		// Iterate over the input slice, casting each element to an integer
		for i, item := range v {
			uids[i] = item.(int)
		}
		// Return a pointer to the resulting slice of integers
		return &uids
	// Handle the case where the value is a schema.Set
	case *schema.Set:
		// Convert the set to a list of interfaces
		setList := v.List()
		// Create a slice of integers with the same length as the set list
		uids := make([]int, len(setList))
		// Iterate over the set list, extracting the "id" field from each map and casting it to an integer
		for i, item := range setList {
			m := item.(map[string]interface{})
			uids[i] = m["id"].(int)
		}
		// Return a pointer to the resulting slice of integers
		return &uids
	// Handle the default case where the type is not recognized
	default:
		// Return nil if the type is not handled
		return nil
	}
}

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

// InflateObjectWithID takes a slice of ObjectWithID and converts it into a slice of maps,
// where each map contains the ID of an ObjectWithID. If the input slice is nil, it returns an empty slice.
func InflateObjectWithID(arr []ObjectWithID) []interface{} {
	if arr != nil {
		// Create an empty slice to hold the converted items
		final := make([]interface{}, 0)

		// Iterate over each item in the input slice
		for _, item := range arr {
			// Create a map to hold the item data
			it := make(map[string]interface{})

			// Add the ID to the map
			it["id"] = item.ID

			// Append the map to the final slice
			final = append(final, it)
		}

		// Return the final slice of maps
		return final
	}

	// Return an empty slice if the input is nil
	return make([]interface{}, 0)
}

// InflateSingleObjectWithID takes a pointer to an ObjectWithID and returns its ID.
// If the input is nil, it returns nil.
func InflateSingleObjectWithID(single *ObjectWithID) interface{} {
	if single != nil {
		// Return the ID of the non-nil ObjectWithID
		return single.ID
	}

	// Return nil if the input is nil
	return nil
}

// InflateArrayOfIDs - Transforms an array of IDs into a map with an "id" key
func InflateArrayOfIDs(arr []int) []interface{} {
	if arr != nil {
		final := make([]interface{}, 0)

		for _, item := range arr {
			it := make(map[string]interface{})

			it["id"] = item

			final = append(final, it)
		}

		return final
	}

	return make([]interface{}, 0)
}

func InflateTags(arr []Tag) map[string]string {
	if arr != nil {
		final := make(map[string]string, 0)

		for _, item := range arr {
			final[item.Key] = item.Value
		}

		return final
	}

	return nil
}

// FieldsChanged -
func FieldsChanged(iOld interface{}, iNew interface{}, fields []string) (map[string]interface{}, string, bool) {
	mOld := iOld.(map[string]interface{})
	mNew := iNew.(map[string]interface{})

	for _, v := range fields {
		if mNew[v] != mOld[v] {
			return mNew, v, true
		}
	}

	return mNew, "", false
}

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

// AssociationChanged returns arrays of which values to change.
// The fields needs to be at the top level.
func AssociationChanged(d *schema.ResourceData, fieldname string) ([]int, []int, bool, error) {
	isChanged := false

	// Get the owner users
	io, in := d.GetChange(fieldname)

	// test for set
	_, isTypeSet := io.(*schema.Set)
	if isTypeSet {
		io = io.(*schema.Set).List()
		in = in.(*schema.Set).List()
	}

	ownerOld := io.([]interface{})
	oldIDs := make([]int, 0)
	for _, item := range ownerOld {
		v, ok := item.(map[string]interface{})
		if ok {
			oldIDs = append(oldIDs, v["id"].(int))
		}
	}
	ownerNew := in.([]interface{})
	newIDs := make([]int, 0)
	for _, item := range ownerNew {
		v, ok := item.(map[string]interface{})
		if ok {
			newIDs = append(newIDs, v["id"].(int))
		}
	}

	arrUserAdd, arrUserRemove, changed := determineAssociations(newIDs, oldIDs)
	if changed {
		isChanged = true
	}

	return arrUserAdd, arrUserRemove, isChanged, nil
}

// AssociationChangedInt returns an int of a value to change.
// The fields needs to be at the top level.
func AssociationChangedInt(d *schema.ResourceData, fieldname string) (*int, *int, bool, error) {
	isChanged := false
	io, in := d.GetChange(fieldname)

	// If the values are not the same, then they changed.
	if in != io {
		isChanged = true

		if in == nil || in == 0 {
			// Either the in value is null which means remove the existing value.
			old := io.(int)
			return nil, &old, isChanged, nil
		}
		// Or the in value is not null which means it should change the
		// existing value.
		newvalue := in.(int)
		return &newvalue, nil, isChanged, nil
	}

	return nil, nil, isChanged, nil
}

// DetermineAssociations will take in a src array (source of truth/repo) and a
// destination array (Kion application) and then return an array of
// associations to add (arrAdd) and then remove (arrRemove).
func determineAssociations(src []int, dest []int) (arrAdd []int, arrRemove []int, isChanged bool) {
	mSrc := makeMapFromArray(src)
	mDest := makeMapFromArray(dest)

	arrAdd = make([]int, 0)
	arrRemove = make([]int, 0)
	isChanged = false

	// Determine which items to add.
	for v := range mSrc {
		if _, found := mDest[v]; !found {
			arrAdd = append(arrAdd, v)
			isChanged = true
		}
	}

	// Determine which items to remove.
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

// GenerateAccTestChecksForResourceOwners returns a list of acceptance test checks for the Owner User & User Group ID
// slices of a given resource.
func GenerateAccTestChecksForResourceOwners(
	resourceType, resourceName string,
	ownerUserIds, ownerUserGroupIds *[]int,
) (funcs []resource.TestCheckFunc) {
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

	return
}

// GenerateOwnerClausesForResourceTest generates a string of owner_users & owner_user_groups clauses to be used in a
// resource declaration for acceptance tests.
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

// TestAccOUGenerateDataSourceDeclarationFilter declares a data source to get an object that matches the name filter
func TestAccOUGenerateDataSourceDeclarationFilter(dataSourceName, localName, name string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {
			filter {
				name = "name"
				values = ["%v"]
			}
		}`, dataSourceName, localName, name,
	)
}

// TestAccOUGenerateDataSourceDeclarationAll declares a data source to get all items
func TestAccOUGenerateDataSourceDeclarationAll(dataSourceName, localName string) string {
	return fmt.Sprintf(`
		data "%v" "%v" {}`, dataSourceName, localName,
	)
}

// PrintHCLConfig prints the generated HCL configuration for unit tests.
func PrintHCLConfig(config string) {
	fmt.Println("Generated HCL configuration:")
	fmt.Println(config)
}
