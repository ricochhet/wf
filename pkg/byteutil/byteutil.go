package byteutil

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ricochhet/pkg/errutil"
)

var (
	ErrInvalidOffsetOrByteRange = errors.New("invalid offset or byte range")
	ErrNoBytes                  = errors.New("no bytes")
)

// WriteBytes writes the specified bytes at the given offset.
func WriteBytes(data []byte, offset int, replace []byte) error {
	if offset < 0 || offset+len(replace) > len(data) {
		return ErrInvalidOffsetOrByteRange
	}

	copy(data[offset:], replace)

	return nil
}

// HexToBytes converts a hexadecimal string to its corresponding byte slice.
func HexToBytes(hex string) ([]byte, error) {
	var data []byte

	for i := 0; i < len(hex); i += 2 {
		var b byte

		_, err := fmt.Sscanf(hex[i:i+2], "%02X", &b)
		if err != nil {
			return nil, errutil.WithFrame(err)
		}

		data = append(data, b)
	}

	return data, nil
}

// Replace replaces occurrences of a pattern in a byte slice.
func Replace(b, old, replacement []byte, i int) []byte {
	var (
		result []byte
		cur    int
	)

	for r := b; len(r) > 0; cur++ {
		bi := bytes.Index(r, old)
		if bi == -1 {
			result = append(result, r...)
			break
		}

		result = append(result, r[:bi]...)

		if i == 0 || cur == i {
			tex := min(len(replacement), len(old))
			result = append(result, replacement[:tex]...)
		} else {
			result = append(result, old...)
		}

		r = r[bi+len(old):]
	}

	return result
}

// IndexAll finds all occurrences of a pattern in a byte slice.
func IndexAll(data, pattern []byte) []int {
	var result []int

	for i := range data {
		if bytes.HasPrefix(data[i:], pattern) {
			result = append(result, i)
		}
	}

	return result
}

// Index finds the index of the first occurrence of pattern in data.
func Index(data, pattern []byte) (int, error) {
	for i := range data[:len(data)-len(pattern)+1] {
		if Match(data[i:i+len(pattern)], pattern) {
			return i, nil
		}
	}

	return -1, ErrNoBytes
}

// Pad pads the bytes to the specified size.
func Pad(data []byte, size int) []byte {
	if len(data) < size {
		ps := size - len(data)
		p := make([]byte, ps)

		return append(data, p...)
	}

	return data
}

// Match matches the bytes in b1 to the bytes in b2.
func Match(b1, b2 []byte) bool {
	for i := range b2 {
		if b1[i] != b2[i] {
			return false
		}
	}

	return true
}
