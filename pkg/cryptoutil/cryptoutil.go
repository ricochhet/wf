package cryptoutil

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"hash"
	"hash/crc32"
	"hash/crc64"
	"io"
	"os"

	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/murmur3"
	"github.com/ricochhet/pkg/strutil"
)

var ErrHashNotEqual = errors.New("hash is not equal")

// NewHash creates a new hash for a file at the specified path.
func NewHash(path string, hash hash.Hash) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", errutil.New("os.Open", err)
	}
	defer file.Close()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", errutil.New("io.Copy", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// Equals validates the hash of a file.
func Equals(path, hash string, t hash.Hash) error {
	hashA, err := NewHash(path, t)
	if err != nil {
		return errutil.New("NewHash", err)
	}

	if hashA != hash {
		return ErrHashNotEqual
	}

	return nil
}

// CatBreadHash returns a catbread hash from the specified name.
func CatBreadHash(name string) uint32 {
	var hash uint32 = 2166136261
	for i := range len(name) {
		hash ^= uint32(name[i])
		hash &= 0x7fffffff // Keep to 31 bits after XOR.
		hash *= 16777619
		hash &= 0x7fffffff // Keep to 31 bits after multiply.
	}

	return hash
}

// NewMD5 returns the MD5 hash of a file.
func NewMD5(path string) (string, error) {
	s, err := NewHash(path, md5.New())
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// NewSHA1 returns the SHA-1 hash of a file.
func NewSHA1(path string) (string, error) {
	s, err := NewHash(path, sha1.New())
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// NewSHA256 returns the SHA-256 hash of a file.
func NewSHA256(path string) (string, error) {
	s, err := NewHash(path, sha256.New())
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// NewSHA512 returns the SHA-512 hash of a file.
func NewSHA512(path string) (string, error) {
	s, err := NewHash(path, sha512.New())
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// NewCRC32 returns the CRC-32 hash of a file.
func NewCRC32(path string) (string, error) {
	s, err := NewHash(path, crc32.New(crc32.IEEETable))
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// NewCRC64 returns the CRC-64 hash of a file.
func NewCRC64(path string) (string, error) {
	s, err := NewHash(path, crc64.New(crc64.MakeTable(crc32.IEEE)))
	if err != nil {
		return "", errutil.WithFrame(err)
	}

	return s, nil
}

// Murmur3X64_128Hash returns the Murmur hash of a file.
func Murmur3X64_128Hash(seed int, str string) uint64 {
	b := murmur3.NewX64_128(seed)
	b.Write(strutil.U8ToU16(str))

	return binary.LittleEndian.Uint64(b.Sum(nil))
}

// Murmur3X86_32Hash returns the Murmur hash of a file.
func Murmur3X86_32Hash(seed int, str string) uint32 {
	b := murmur3.NewX86_32(seed)
	b.Write(strutil.U8ToU16(str))

	return binary.LittleEndian.Uint32(b.Sum(nil))
}

// Murmur3X86_128Hash returns the Murmur hash of a file.
func Murmur3X86_128Hash(seed int, str string) uint32 {
	bytes := murmur3.NewX86_128(seed)
	bytes.Write(strutil.U8ToU16(str))

	return binary.LittleEndian.Uint32(bytes.Sum(nil))
}
