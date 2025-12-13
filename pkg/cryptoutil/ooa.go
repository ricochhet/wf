package cryptoutil

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
)

const (
	OOACipherTag    = "<CipherKey>"
	OOABase64_16Len = 24
)

var OOADlfKey = []byte{
	65, 50, 114, 45, 208, 130, 239, 176, 220, 100, 87, 197, 118, 104, 202, 9,
}

var OOAIV = make([]byte, 16)

var (
	ErrDlfFileNotFound   = errors.New("error: DLF file not found")
	ErrCipherTagNotFound = errors.New("error: Cipher tag not found")
	ErrInvalidBase64Key  = errors.New("error: invalid base64 key")
	ErrInvalidIVSize     = errors.New("error: invalid IV size")
	ErrInvalidBufferSize = errors.New("error: invalid buffer size")
)

// AESEncrypt encrypts the text using the key and returns the encrypted data.
func AESEncrypt(key, text []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, errutil.New("aes.NewCipher", err)
	}

	data := make([]byte, aes.BlockSize+len(text))
	iv := data[:aes.BlockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, errutil.New("io.ReadFull", err)
	}

	m := cipher.NewCBCEncrypter(b, iv)
	m.CryptBlocks(data[aes.BlockSize:], text)

	return data, nil
}

// AESDecrypt decrypts the text using the key and returns the decrypted data.
func AESDecrypt(key, iv, buf []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, errutil.New("aes.NewCipher", err)
	}

	if len(iv) != aes.BlockSize {
		return nil, ErrInvalidIVSize
	}

	if len(buf) < aes.BlockSize {
		return nil, ErrInvalidBufferSize
	}

	m := cipher.NewCBCDecrypter(b, iv)
	m.CryptBlocks(buf, buf)

	return buf, nil
}

// AESDecryptBase64 decrypts the text using the key and returns the decrypted data.
func AESDecryptBase64(kb64 string, iv, buf []byte) error {
	k, err := base64.StdEncoding.DecodeString(kb64)
	if err != nil {
		return ErrInvalidBase64Key
	}

	b, err := aes.NewCipher(k)
	if err != nil {
		return errutil.New("aes.NewCipher", err)
	}

	if len(iv) != aes.BlockSize {
		return ErrInvalidIVSize
	}

	if len(buf) < aes.BlockSize {
		return ErrInvalidBufferSize
	}

	m := cipher.NewCBCDecrypter(b, iv)
	m.CryptBlocks(buf, buf)

	return nil
}

// DecryptDLF decrypts the text using the key and returns the decrypted data.
func DecryptDLF(data []byte) ([]byte, error) {
	d, err := AESDecrypt(OOADlfKey, data[0x41:], []byte{0})
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return d, nil
}

// DLFAuto returns the decrypted data for the given CID.
func DLFAuto(cid string) ([]byte, error) {
	paths := []string{
		cid + ".dlf",
		cid + "_cached.dlf",
	}

	for _, path := range paths {
		data, err := fsutil.ReadBytes(path)
		if err == nil {
			return DecryptDLF(data)
		}
	}

	return nil, ErrDlfFileNotFound
}

// CipherTag decodes the cipher tag from the given data.
func CipherTag(dlf []byte) ([]byte, error) {
	data := string(dlf)
	pos := strings.Index(data, OOACipherTag)

	if pos == -1 {
		return nil, ErrCipherTagNotFound
	}

	pos += len(OOACipherTag)
	b64 := data[pos : pos+OOABase64_16Len]

	decode, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, errutil.New("base64.StdEncoding.DecodeString", err)
	}

	if len(decode) > 16 {
		decode = decode[:16]
	}

	return decode, nil
}

// OOAHash returns the OOA hash from the given data.
func OOAHash(data []byte) []byte {
	if len(data) < 0x3E {
		return nil
	}

	return data[0x2A:0x3E]
}
