package embedutil

import (
	"embed"
	"encoding/json"
	"path/filepath"
	"strings"

	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
)

type EmbeddedFileSystem struct {
	Initial string
	FS      embed.FS
}

// List return a list of files within the embedded fs, calling the list function in return.
func (e *EmbeddedFileSystem) List(pattern string, list func([]File) error) error {
	files, err := WalkDir(e.FS, pattern)
	if err != nil {
		return errutil.WithFrame(err)
	}

	return list(files)
}

// Dump return a []byte within the embedded fs, calling the dump function in return.
func (e *EmbeddedFileSystem) Dump(pattern, name string, dump func(File, []byte) error) error {
	files, err := WalkDir(e.FS, pattern)
	if err != nil {
		return errutil.New("NewFileGetter", err)
	}

	for _, file := range files {
		if !strings.Contains(file.Path, name) {
			continue
		}

		data, err := e.FilenameToBytes(strings.TrimPrefix(file.Path, e.Initial))
		if err != nil {
			return errutil.New("FilenameToByte", err)
		}

		if err := dump(file, data); err != nil {
			return errutil.New("dump", err)
		}
	}

	return nil
}

// FallbackToBytes reads the filename from a path, falling back to embed if it does not exist.
func (e *EmbeddedFileSystem) FallbackToBytes(name string) ([]byte, error) {
	path := filepath.Join(e.Initial, name)
	if fsutil.Exists(path) {
		return fsutil.ReadBytes(path)
	}

	return e.FilenameToBytes(name)
}

// BytesToMap converts a byte array into a map[string]any.
func (e *EmbeddedFileSystem) BytesToMap(data []byte) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, errutil.WithFrame(err)
	}

	return m, nil
}

// FilenameToMap converts a file in the embedded filesystem into a map[string]any.
func (e *EmbeddedFileSystem) FilenameToMap(name string) (map[string]any, error) {
	b, err := e.FS.ReadFile(e.Initial + name)
	if err != nil {
		return nil, errutil.New("FS.ReadFile", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errutil.New("json.Unmarshal", err)
	}

	return m, nil
}

// FilenameToBytes converts a file in the embedded filesystem into an array of bytes.
func (e *EmbeddedFileSystem) FilenameToBytes(name string) ([]byte, error) {
	b, err := e.FS.ReadFile(e.Initial + name)
	if err != nil {
		return nil, errutil.New("FS.ReadFile", err)
	}

	var r json.RawMessage
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, errutil.New("json.Unmarshal", err)
	}

	return b, nil
}
