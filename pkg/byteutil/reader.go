package byteutil

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/ricochhet/pkg/errutil"
)

type Reader struct {
	file *os.File
}

// NewReader creates a new reader for the given file.
func NewReader(name string) (*Reader, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Reader{file}, nil
}

// IsValid returns true if the reader is valid.
func (r *Reader) IsValid() bool {
	return r.file != nil
}

// ReadUInt32 reads a 32-bit unsigned integer from the file.
func (r *Reader) ReadUInt32() (uint32, error) {
	var v uint32

	err := binary.Read(r.file, binary.LittleEndian, &v)

	return v, errutil.WithFrame(err)
}

// ReadUInt64 reads a 64-bit unsigned integer from the file.
func (r *Reader) ReadUInt64() (uint64, error) {
	var v uint64

	err := binary.Read(r.file, binary.LittleEndian, &v)

	return v, errutil.WithFrame(err)
}

// Read reads data from the reader.
func (r *Reader) Read(data []byte) (int, error) {
	return r.file.Read(data)
}

// ReadChar reads a single character from the reader.
func (r *Reader) ReadChar() (byte, error) {
	var v byte

	err := binary.Read(r.file, binary.LittleEndian, &v)

	return v, errutil.WithFrame(err)
}

// Seek sets the position of the reader.
func (r *Reader) Seek(position int64, whence int) (int64, error) {
	return r.file.Seek(position, whence)
}

// SeekFromBeginning set the position of the reader to the beginning of the file.
func (r *Reader) SeekFromBeginning(position int64) (int64, error) {
	return r.file.Seek(position, io.SeekStart)
}

// SeekFromEnd sets the position of the reader to the end of the file.
func (r *Reader) SeekFromEnd(position int64) (int64, error) {
	return r.file.Seek(position, io.SeekEnd)
}

// SeekFromCurrent sets the position of the reader to a specific position in the file.
func (r *Reader) SeekFromCurrent(position int64) (int64, error) {
	return r.file.Seek(position, io.SeekCurrent)
}

// Position gets the current position of the reader.
func (r *Reader) Position() (int64, error) {
	return r.file.Seek(0, io.SeekCurrent)
}

// Size gets the size of the file.
func (r *Reader) Size() (size int64, err error) {
	cur, err := r.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, errutil.New("file.Seek", err)
	}

	defer func() {
		if _, seekErr := r.file.Seek(cur, io.SeekStart); err != nil {
			size = 0
			err = errutil.New("seekErr", seekErr)
		}
	}()

	size, err = r.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, errutil.New("file.Seek", err)
	}

	return size, nil
}

// Close closes the reader.
func (r *Reader) Close() error {
	return r.file.Close()
}
