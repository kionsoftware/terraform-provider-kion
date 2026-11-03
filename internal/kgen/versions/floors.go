package versions

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"sort"

	"gopkg.in/yaml.v3"
)

// The SDK ships one package per Kion minor, so derivation can only say
// "3.NN.0". A resource that first appeared in a patch release needs the patch
// stated: resource_floors in codegen/attr_version_overrides.yaml raises the
// derived minimum for the resource and its data source. The same code serves
// cmd/kversions, so both outputs agree.

var floorRE = regexp.MustCompile(`^3\.\d+\.\d+$`)

// decodeAttrOverrides decodes attr_version_overrides.yaml strictly: an unknown
// key (a typo such as resource_floor) is an error, not a silently ignored entry.
func decodeAttrOverrides(b []byte) (attrOverridesFile, error) {
	var f attrOverridesFile
	dec := yaml.NewDecoder(bytes.NewReader(b))
	dec.KnownFields(true)
	if err := dec.Decode(&f); err != nil && !errors.Is(err, io.EOF) {
		return f, err
	}
	return f, nil
}

// LoadFloors reads attr_version_overrides.yaml through readFile and returns its
// resource_floors, validated against the resource and data source names. Only a
// missing file means "no floors"; any other read error is returned.
func LoadFloors(readFile func(string) ([]byte, error), path string, resources, dataSources []string) (map[string]string, error) {
	b, err := readFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	f, err := decodeAttrOverrides(b)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	for _, name := range sortedResourceNames(f.ResourceFloors) {
		if v := f.ResourceFloors[name]; !floorRE.MatchString(v) {
			return nil, fmt.Errorf("resource_floors: %s: %q is not a 3.N.P version", name, v)
		}
	}
	if err := checkFloorTargets(f.ResourceFloors, resources, dataSources); err != nil {
		return nil, err
	}
	if f.ResourceFloors == nil {
		return map[string]string{}, nil
	}
	return f.ResourceFloors, nil
}

// DefState is what derivation found for one config entry.
type DefState int

// What derivation found for an entry; see ApplyFloor.
const (
	FloorNoOp       DefState = iota // entry has no create or read operation
	FloorUnresolved                 // defining op is in no tracked SDK version
	FloorResolved                   // a range was derived
)

// ApplyFloor is the per-entry floor decision both generators share. With no
// floor it returns (derivedMin, false, nil). A floor on an entry with nothing
// derived is an error. On a resolved entry it returns the raised minimum and
// emit=true, so an entry that needed no gate still gets one.
func ApplyFloor(name, section, floor string, st DefState, defOp, derivedMin, derivedMax string) (string, bool, error) {
	if floor == "" {
		return derivedMin, false, nil
	}
	switch st {
	case FloorNoOp:
		return "", false, fmt.Errorf("resource_floors: %s (%s): %s has no create or read operation to derive a range from", name, floor, section)
	case FloorUnresolved:
		return "", false, fmt.Errorf("resource_floors: %s (%s): %s defining op %q is in no tracked SDK version", name, floor, section, defOp)
	}
	m, err := RaiseMin(name, derivedMin, derivedMax, floor)
	if err != nil {
		return "", false, err
	}
	return m, true, nil
}

// RaiseMin applies floor to a derived minimum. A floor can only tighten: one
// below the derived minimum is an error, as is one above a bounded maximum.
// An empty floor returns derivedMin unchanged.
func RaiseMin(name, derivedMin, derivedMax, floor string) (string, error) {
	if floor == "" {
		return derivedMin, nil
	}
	if versionLess(floor, derivedMin) {
		return "", fmt.Errorf("resource_floors: %s: floor %s is below the derived minimum %s; a floor can only tighten", name, floor, derivedMin)
	}
	if derivedMax != "" && !versionLess(floor, derivedMax) {
		return "", fmt.Errorf("resource_floors: %s: floor %s is not below the derived maximum %s", name, floor, derivedMax)
	}
	return floor, nil
}

// checkFloorTargets fails when a floor names something that is neither a
// resource nor a data source: a floor nothing reads is worse than none.
func checkFloorTargets(floors map[string]string, resources, dataSources []string) error {
	known := make(map[string]bool, len(resources)+len(dataSources))
	for _, n := range resources {
		known[n] = true
	}
	for _, n := range dataSources {
		known[n] = true
	}
	for _, name := range sortedResourceNames(floors) {
		if !known[name] {
			return fmt.Errorf("resource_floors: %s (%s) is neither a resource nor a data source in the generator config", name, floors[name])
		}
	}
	return nil
}

// SortedKeys returns the map's keys in order.
func SortedKeys[T any](m map[string]T) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
