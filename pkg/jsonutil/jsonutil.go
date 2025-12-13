package jsonutil

import (
	"fmt"
	"strings"

	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/maputil"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// ResultAsArray gets a json result as a slice of gjson.Result.
func ResultAsArray(data []byte, name string, index int) ([]gjson.Result, error) {
	t, err := Result(data, name, index)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return t.Array(), nil
}

// Result gets a result slice at the specified index.
func Result(data []byte, path string, index int) (gjson.Result, error) {
	array := gjson.ParseBytes(data).Array()
	if index < 0 || index >= len(array) {
		return gjson.Result{}, errutil.WithFramef("invalid index: %d", index)
	}

	target := array[index]

	return target.Get(path), nil
}

// ResultToSlice converts a gjson.Result slice to a string slice.
func ResultToSlice(m []gjson.Result) []string {
	result := []string{}
	for _, item := range m {
		result = append(result, item.Raw)
	}

	return result
}

// UniqueToSlice converts a map with a gjson path to a string slice.
func UniqueToSlice(m []string, path string) []string {
	result := make(map[string]string, len(m))
	for _, raw := range m {
		result[gjson.Get(raw, path).String()] = raw
	}

	return maputil.MapToSlice(result)
}

// UniqueToMap converts a slice with a gjson path to a map.
func UniqueToMap(m []string, path string) map[string]string {
	result := make(map[string]string, len(m))
	for _, raw := range m {
		result[gjson.Get(raw, path).String()] = raw
	}

	return result
}

// SetSliceInRawBytes sets a slice to the input json bytes at the specified index.
func SetSliceInRawBytes(input []byte, path string, elems []string, index int) ([]byte, error) {
	return sjson.SetRawBytes(
		input,
		fmt.Sprintf("%d.%s", index, path),
		[]byte("["+strings.Join(elems, ",")+"]"),
	)
}

// SetFieldInRawBytes sets a field to the input json bytes at the specified index.
func SetFieldInRawBytes(input []byte, path, elem string, index int) ([]byte, error) {
	return sjson.SetRawBytes(input, fmt.Sprintf("%d.%s", index, path), []byte(elem))
}

// SetFieldInRawBytes sets a field to the input json string at the specified index.
func SetFieldInBytes[T any](input, path string, elem T, index int) (string, error) {
	return sjson.Set(input, fmt.Sprintf("%d.%s", index, path), elem)
}
