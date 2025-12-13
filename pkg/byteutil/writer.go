package byteutil

import (
	"encoding/binary"
	"io"
	"os"

	"github.com/ricochhet/pkg/errutil"
)

type Writer struct {
	file *os.File
}

type FileEntry struct {
	FileName      string
	FileNameLower uint32
	FileNameUpper uint32
	Offset        uint64
	UncompSize    uint64
}

type DataEntry struct {
	Hash     uint32
	FileName string
}

// FindByHash returns the first entry in data with a matching hash.
func FindByHash(data []DataEntry, hash uint32) *DataEntry {
	for _, d := range data {
		if d.Hash == hash {
			return &d
		}
	}

	return nil
}

// FindByFileName returns the first entry in data with a matching file name.
func FindByFileName(data []DataEntry, name string) *DataEntry {
	for _, d := range data {
		if d.FileName == name {
			return &d
		}
	}

	return nil
}

// NewWriter creates a new writer for the given file name.
func NewWriter(name string, appendMode bool) (*Writer, error) {
	var file *os.File

	var err error

	if appendMode {
		file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	} else {
		file, err = os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	}

	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Writer{file}, nil
}

// WriteUInt32 writes a 32-bit unsigned integer to the file.
func (w *Writer) WriteUInt32(value uint32) error {
	return binary.Write(w.file, binary.LittleEndian, value)
}

// WriteUInt64 writes a 64-bit unsigned integer to the file.
func (w *Writer) WriteUInt64(value uint64) error {
	return binary.Write(w.file, binary.LittleEndian, value)
}

// Write writes data to the writer.
func (w *Writer) Write(data []byte) (int, error) {
	return w.file.Write(data)
}

// WriteChar writes a single character to the writer.
func (w *Writer) WriteChar(data string) (int, error) {
	return w.file.WriteString(data)
}

// Seek sets the position of the writer.
func (w *Writer) Seek(position int64, whence int) (int64, error) {
	return w.file.Seek(position, whence)
}

// SeekFromBeginning sets the position of the reader to the beginning of the file.
func (w *Writer) SeekFromBeginning(position int64) (int64, error) {
	return w.file.Seek(position, io.SeekStart)
}

// SeekFromEnd sets the position of the reader to the end of the file.
func (w *Writer) SeekFromEnd(position int64) (int64, error) {
	return w.file.Seek(position, io.SeekEnd)
}

// SeekFromCurrent sets the position of the reader to a specific position in the file.
func (w *Writer) SeekFromCurrent(position int64) (int64, error) {
	return w.file.Seek(position, io.SeekCurrent)
}

// Position gets the current position of the writer.
func (w *Writer) Position() (int64, error) {
	return w.file.Seek(0, io.SeekCurrent)
}

// Size gets the size of the file.
func (w *Writer) Size() (size int64, err error) {
	cur, err := w.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, errutil.New("file.Seek", err)
	}

	defer func() {
		if _, seekErr := w.file.Seek(cur, io.SeekStart); err != nil {
			size = 0
			err = errutil.New("seekErr", seekErr)
		}
	}()

	size, err = w.file.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, errutil.New("file.Seek", err)
	}

	return size, nil
}

// Close closes the file.
func (w *Writer) Close() error {
	return w.file.Close()
}
