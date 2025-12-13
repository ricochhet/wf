package strutil

import (
	"bytes"
	"unicode/utf16"
	"unsafe"
)

// ReadU16String read a section of []byte as a UTF16 string.
func ReadU16String(data []byte, start, end int) string {
	var cid string

	if end > len(data) {
		end = len(data)
	}

	raw := data[start:end]
	u16 := ((*[1 << 30]uint16)(unsafe.Pointer(&raw[0])))[:len(raw)/2]
	ind := -1

	for i, c := range u16 {
		if c == 0 {
			ind = i
			break
		}
	}

	if ind != -1 {
		cid = string(utf16.Decode(u16[:ind]))
	} else {
		cid = string(utf16.Decode(u16))
	}

	return cid
}

// StringToBytes converts a string to bytes with padding.
func StringToBytes(str string, pad int) []byte {
	tmp := []byte(str)
	tmp = append(tmp, bytes.Repeat([]byte{0}, pad-(len(tmp)%pad))...)

	return tmp
}

// U8ToU16 converts an UTF-8 string to a UTF-16 array.
func U8ToU16(u8 string) []byte {
	b := []byte(u8)
	r := utf16.Encode([]rune(string(b)))
	u16b := make([]byte, len(r)*2)

	for i, r := range r {
		u16b[i*2] = byte(r)
		u16b[i*2+1] = byte(r >> 8)
	}

	return u16b
}
