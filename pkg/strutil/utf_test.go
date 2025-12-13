package strutil_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/ricochhet/pkg/strutil"
)

var ErrUnexpectedBytes = errors.New("unexpected bytes")

func TestUtf8ToUtf16(t *testing.T) {
	t.Parallel()

	b := strutil.U8ToU16("aaabbbccc")
	o := []byte{97, 0, 97, 0, 97, 0, 98, 0, 98, 0, 98, 0, 99, 0, 99, 0, 99, 0}

	if !bytes.Equal(b, o) {
		t.Fatal(ErrUnexpectedBytes)
	}
}
