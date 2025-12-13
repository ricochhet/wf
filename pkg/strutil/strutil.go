package strutil

import (
	"strings"
)

// Format replacess all occurrences of replacements[i] in str.
func Format(str string, replacements map[string]string) string {
	for k, v := range replacements {
		str = strings.ReplaceAll(str, k, v)
	}

	return str
}

// StartsWithAny checks if a string starts with any prefixes in the slice.
func StartsWithAny(str string, prefixes []string) bool {
	for _, s := range prefixes {
		if strings.HasPrefix(str, s) {
			return true
		}
	}

	return false
}

// EndsWithAny checks if a string ends with any suffixes in the slice.
func EndsWithAny(str string, suffixes []string) bool {
	for _, s := range suffixes {
		if strings.HasSuffix(str, s) {
			return true
		}
	}

	return false
}
