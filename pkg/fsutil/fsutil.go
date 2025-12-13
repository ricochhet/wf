package fsutil

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/otiai10/copy"
	"github.com/ricochhet/pkg/errutil"
)

type Match int

const (
	File Match = iota
	Directory
	Any
)

// Normalize replaces all backslashes with forward slashes in a string.
func Normalize(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// ToNormalized returns a string slice of a path.
func ToNormalized(path string) []string {
	return strings.Split(Normalize(path), "/")
}

// ToRelative returns the set of paths as relative paths.
func ToRelative(paths ...string) string {
	result := "./" + paths[0]
	for _, dir := range paths[1:] {
		result = path.Join(result, dir)
	}

	return result
}

// TrimPath trims the prefixes and suffixes from path.
func TrimPath(path string) string {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, ".\\") {
		return path[2:]
	} else if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		return path[1:]
	}

	if strings.HasSuffix(path, "/.") || strings.HasSuffix(path, "\\.") {
		return path[:len(path)-2]
	} else if strings.HasSuffix(path, "/") || strings.HasSuffix(path, "\\") {
		return path[:len(path)-1]
	}

	return path
}

// FilenameToMap reads a file and convert it into a map[string]any.
func FilenameToMap(initial, name string) (map[string]any, error) {
	b, err := os.ReadFile(initial + name)
	if err != nil {
		return nil, errutil.New("os.ReadFile", err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, errutil.New("json.Unmarshal", err)
	}

	return m, nil
}

// FilenameToBytes reads a file return it as an array of bytes.
func FilenameToBytes(initial, name string) ([]byte, error) {
	b, err := os.ReadFile(initial + name)
	if err != nil {
		return nil, errutil.New("os.ReadFile", err)
	}

	var r json.RawMessage
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, errutil.New("json.Unmarshal", err)
	}

	return b, nil
}

// WriteBytes writes to a file, creating the full path if necessary.
func WriteBytes(name string, data []byte, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(name), perm); err != nil {
		return errutil.New("os.MkDirAll", err)
	}

	if err := os.WriteFile(name, data, perm); err != nil {
		return errutil.New("os.WriteFile", err)
	}

	return nil
}

// WriteString writes string lines to a file.
func WriteString(file *os.File, lines []string) error {
	for _, lines := range lines {
		if _, err := file.WriteString(lines); err != nil {
			return errutil.WithFrame(err)
		}
	}

	return nil
}

// Overwrite truncates a file to the specified size, seeking to the specified value.
func Overwrite(file *os.File, size, offset int64, whence int) error {
	if err := file.Truncate(size); err != nil {
		return errutil.New("file.Truncate", err)
	}

	if _, err := file.Seek(offset, whence); err != nil {
		return errutil.New("file.Seek", err)
	}

	return nil
}

// ReadBytes reads bytes from a file.
func ReadBytes(name string) ([]byte, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return data, errutil.WithFrame(err)
	}

	return data, nil
}

// CombineEnviron combines the given envs with the env by name.
func CombineEnviron(name string, envs []string) string {
	separator := string(filepath.ListSeparator)
	existing := os.Getenv(name)
	joined := strings.Join(envs, separator)

	if existing != "" {
		return name + "=" + joined + separator + existing
	}

	return name + "=" + joined
}

// URLFilename gets the filename of a file from a url.
func URLFilename(v string) (string, error) {
	u, err := url.Parse(v)
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return path.Base(u.Path), nil
}

// Exists returns true if a file exists.
func Exists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

// Filename returns the filename of a path without the extension.
func Filename(name string) string {
	return strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
}

// FromCwd combines paths starting from the current working directory.
func FromCwd(pathA ...string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	path := append([]string{wd}, pathA...)

	return filepath.Join(path...), nil
}

// RemoveDirectory removes all parts of a directory, skipping deletion if skip is truthy.
func RemoveDirectory(name string, skip func(string) bool) error {
	return filepath.Walk(name, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errutil.WithFrame(err)
		}

		if path == name {
			return nil
		}

		if !info.IsDir() {
			if !skip(path) {
				return os.Remove(path)
			}

			return nil
		}

		return nil
	})
}

// RemoveAll removes all parts of a path.
func RemoveAll(name string) error {
	err := os.RemoveAll(name)
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// RemoveAllEmpty removes all empty directories in a directory.
func RemoveAllEmpty(name string) error {
	var directories []string

	err := filepath.WalkDir(name, func(path string, directory os.DirEntry, err error) error {
		if err != nil {
			return errutil.New("filepath.WalkDir", err)
		}

		if directory.IsDir() && path != name {
			directories = append(directories, path)
		}

		return nil
	})
	if err != nil {
		return errutil.WithFrame(err)
	}

	for i := len(directories) - 1; i >= 0; i-- {
		d := directories[i]

		empty, err := IsEmpty(d)
		if err != nil {
			continue
		}

		if empty {
			err = os.Remove(d)
			if err != nil {
				continue
			}
		}
	}

	return nil
}

// IsEmpty checks if a path is empty.
func IsEmpty(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, errutil.New("os.Open", err)
	}
	defer file.Close()

	_, err = file.Readdir(1)
	if err == nil {
		return false, nil
	}

	if errors.Is(err, os.ErrNotExist) || err.Error() == "EOF" {
		return true, nil
	}

	return false, errutil.WithFrame(err)
}

// Copy copies from path1 to path2.
func Copy(path1, path2 string, opts ...copy.Options) error {
	if err := copy.Copy(path1, path2, opts...); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// CopyFile copies a file from path1 to path2 using io.Copy.
func CopyFile(path1, path2 string) (int64, error) {
	src, err := os.Open(path1)
	if err != nil {
		return 0, errutil.New("os.Open", err)
	}
	defer src.Close()

	dest, err := os.Create(path2)
	if err != nil {
		return 0, errutil.New("os.Create", err)
	}
	defer dest.Close()

	return io.Copy(dest, src)
}

// Scan scans a file and return all lines as a slice.
func Scan(scanner *bufio.Scanner) ([]string, error) {
	lines := []string{}

	for scanner.Scan() {
		if len(scanner.Text()) == 0 {
			continue
		}

		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, errutil.WithFrame(err)
	}

	return lines, nil
}

// Sort sorts paths alphabetically.
func Sort(paths []string) []string {
	sort.Slice(paths, func(i, j int) bool {
		p1 := filepath.Dir(paths[i])
		p2 := filepath.Dir(paths[j])

		if p1 == p2 {
			return filepath.Base(paths[i]) < filepath.Base(paths[j])
		}

		return p1 < p2
	})

	return paths
}

// Find finds everything that matches the type at the path.
func Find(path string, match Match) []string {
	result := []string{}

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errutil.WithFrame(err)
		}

		switch match {
		case File:
			if !info.IsDir() {
				result = append(result, path)
			}
		case Directory:
			if info.IsDir() {
				result = append(result, path)
			}
		case Any:
			result = append(result, path)
		}

		return nil
	})
	if err != nil {
		return []string{}
	}

	return Sort(result)
}

// FindRoot finds everything that matches the type at the root path.
func FindRoot(path string, match Match) ([]string, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	var directories []string

	for _, entry := range entries {
		switch match {
		case File:
			if !entry.IsDir() {
				directories = append(directories, entry.Name())
			}
		case Directory:
			if entry.IsDir() {
				directories = append(directories, entry.Name())
			}
		case Any:
			directories = append(directories, entry.Name())
		}
	}

	return directories, nil
}

// Depth returns the depth of a file path.
func Depth(path string) int {
	if path == "." || path == "" {
		return 0
	}

	cur := 0

	for {
		dir, file := filepath.Split(path)
		if file != "" {
			cur++
		}

		if dir == "" || dir == "/" || dir == "." {
			break
		}

		path = dir[:len(dir)-1]
	}

	return cur
}
