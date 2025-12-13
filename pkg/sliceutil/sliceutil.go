package sliceutil

import "slices"

// Matches returns the number of matches between s1 and s2.
func Matches(s1, s2 []string) int {
	seen := make(map[string]struct{}, len(s2))
	for _, e := range s2 {
		seen[e] = struct{}{}
	}

	var i int

	for _, part := range s1 {
		if _, found := seen[part]; found {
			i++
		}
	}

	return i
}

// Replace replaces the first occurrence of s1 in s with s2.
func Replace[T comparable](s, s1, s2 []T) []T {
	for i := 0; i <= len(s)-len(s1); i++ {
		if slices.Equal(s[i:i+len(s1)], s1) {
			return slices.Concat(s[:i], append(s2, s[i+len(s1):]...))
		}
	}

	return s
}

// Subslice checks if s1 exists in s2.
func Subslice[T comparable](s1, s2 []T) bool {
	if len(s2) == 0 || len(s2) > len(s1) {
		return false
	}

	for i := 0; i <= len(s1)-len(s2); i++ {
		match := true

		for j := range s2 {
			if s1[i+j] != s2[j] {
				match = false
				break
			}
		}

		if match {
			return true
		}
	}

	return false
}

// Move moves element from its current position to index.
func Move[T comparable](s []T, elem T, index int) []T {
	cur := -1

	for i, s := range s {
		if s == elem {
			cur = i
			break
		}
	}

	if cur == -1 {
		return s
	}

	s = slices.Delete(s, cur, cur+1)
	if index >= len(s) {
		s = append(s, elem)
	} else {
		s = append(s[:index], append([]T{elem}, s[index:]...)...)
	}

	return s
}
