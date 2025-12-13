package maputil

import (
	"encoding/json"
	"maps"
	"reflect"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/ricochhet/pkg/errutil"
)

// SliceToKeyValuePairs convert a slice of string key=value pairs to a map[key]value.
func SliceToKeyValuePairs(kv []string) (map[string]string, error) {
	m := map[string]string{}

	for _, line := range kv {
		kvp := strings.SplitN(line, "=", 2)
		if len(kvp) != 2 {
			continue
		}

		m[kvp[0]] = kvp[1]
	}

	return m, nil
}

// MergeMap merges m1 into m2, with m2 overwriting any keys in m1.
func MergeMap[K comparable, V any](m1, m2 map[K]V) map[K]V {
	m := make(map[K]V, len(m1)+len(m2))

	maps.Copy(m, m1)
	maps.Copy(m, m2)

	return m
}

// MapToSlice converts a map to a string slice.
func MapToSlice(m map[string]string) []string {
	result := make([]string, 0, len(m))
	for _, raw := range m {
		result = append(result, raw)
	}

	return result
}

// BytesToMap converts a byte slice to a map.
func BytesToMap(data []byte) (map[string]any, error) {
	var b map[string]any
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, errutil.WithFrame(err)
	}

	return b, nil
}

// Merge merges T, replacing m1 with m2.
func Merge[T any](m1, m2 T, tagName string, weaklyTypedInput bool) T {
	map1 := StructToMap(m1, tagName, weaklyTypedInput)
	map2 := StructToMap(m2, tagName, weaklyTypedInput)

	for k, v := range map2 {
		if IsZero(v) {
			delete(map2, k)
		}
	}

	maps.Copy(map1, map2)

	var merged T

	_ = mapstructure.Decode(map1, &merged)

	return merged
}

// AppendNewByKey appends elements from slice m2 to slice m1.
func AppendNewByKey[T any, K comparable](m1, m2 []T, key func(T) K) []T {
	seen := make(map[K]struct{}, len(m1)+len(m2))
	result := []T{}

	for _, v := range m1 {
		k := key(v)
		seen[k] = struct{}{}

		result = append(result, v)
	}

	for _, v := range m2 {
		k := key(v)
		if _, exists := seen[k]; !exists {
			result = append(result, v)
			seen[k] = struct{}{}
		}
	}

	return result
}

// AppendOverwriteByKey merges slice m1 with slice m2.
func AppendOverwriteByKey[T any, K comparable](m1, m2 []T, key func(T) K) []T {
	seen := make(map[K]T, len(m1)+len(m2))
	order := []K{}

	for _, v := range m1 {
		k := key(v)
		seen[k] = v
		order = append(order, k)
	}

	for _, v := range m2 {
		k := key(v)
		if _, ok := seen[k]; !ok {
			order = append(order, k)
		}

		seen[k] = v
	}

	result := make([]T, 0, len(seen))
	for _, k := range order {
		result = append(result, seen[k])
	}

	return result
}

// StructToMap converts v to map[string]any.
func StructToMap(v any, tagName string, weaklyTypedInput bool) map[string]any {
	var out map[string]any

	cfg := &mapstructure.DecoderConfig{
		Result:           &out,
		TagName:          tagName,
		WeaklyTypedInput: weaklyTypedInput,
	}
	dec, _ := mapstructure.NewDecoder(cfg)
	_ = dec.Decode(v)

	return out
}

// IsZero checks if the value of v is zero.
func IsZero(v any) bool {
	return reflect.ValueOf(v).IsZero()
}
