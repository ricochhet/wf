package dlutil

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/avast/retry-go"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
)

var (
	ErrDownloadURLEmpty  = errors.New("download url is empty")
	ErrDownloadPathEmpty = errors.New("download path is empty")
	ErrDownloadNameEmpty = errors.New("download name is empty")
	ErrFileHashNoMatch   = errors.New("file hash does not match")
)

type Messenger struct {
	Start func(string)
}

type Download struct {
	URL       string
	Directory string
	Filename  string
	SHA256    string
}

var retryOpts = []retry.Option{
	retry.Attempts(5),
	retry.LastErrorOnly(true),
	retry.OnRetry(func(n uint, err error) {
		logutil.Infof(os.Stdout, "Retry %d: %v\n", n, err)
	}),
}

// RetryOpts sets the options for the downloader.
func RetryOpts(opts []retry.Option) {
	retryOpts = opts
}

// NewDefaultMessenger creates a default Messenger.
func NewDefaultMessenger() Messenger {
	return Messenger{
		Start: func(name string) {
			logutil.Infof(os.Stdout, "Downloading: %s\n", name)
		},
	}
}

// NewDefaultValidator creates a default validator function using SHA256.
func (d *Download) NewDefaultValidator(path string) error {
	if fsutil.Exists(path) {
		if d.SHA256 == "" {
			return nil
		}

		b, err := os.ReadFile(path)
		if err == nil {
			sha := sha256.New()
			sha.Write(b)
			shasum := hex.EncodeToString(sha.Sum(nil))

			if strings.ToLower(d.SHA256) == shasum {
				logutil.Infof(os.Stdout, "Ok: %s\n", d.Filename)
				return nil
			}
		}
	}

	return ErrFileHashNoMatch
}

// Download downloads a file.
func (d *Download) Download(
	ctx context.Context,
	messenger Messenger,
	validator func(string) error,
) error {
	if err := d.ensureDownloadParams(); err != nil {
		return errutil.New("validateParams", err)
	}

	path := filepath.Join(d.Directory, d.Filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return errutil.New("os.MkdirAll", err)
	}

	return retry.Do(func() error {
		if validator != nil {
			if err := validator(path); err == nil {
				return nil
			}
		}

		messenger.Start(d.Filename)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
		if err != nil {
			return errutil.New("http.NewRequestWithContext", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return errutil.New("http.DefaultClient.Do", err)
		}
		defer res.Body.Close()

		tmp := path + ".tmp"

		file, err := os.Create(tmp)
		if err != nil {
			return errutil.New("os.Create", err)
		}

		if err := d.writeResp(res, file, validator == nil || d.SHA256 == ""); err != nil {
			file.Close()
			os.Remove(tmp)

			return errutil.New("writeResp", err)
		}

		if err := file.Close(); err != nil {
			os.Remove(tmp)
			return errutil.New("file.Close", err)
		}

		return os.Rename(tmp, path)
	}, retryOpts...)
}

// writeResp writes the response to the given file path.
func (d *Download) writeResp(resp *http.Response, flags *os.File, skip bool) error {
	sha := sha256.New()
	buf := make([]byte, 1<<20)

	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			return errutil.New("resp.Body.Read", err)
		}

		if n == 0 {
			break
		}

		if _, err := flags.Write(buf[:n]); err != nil {
			return errutil.New("flags.Write", err)
		}

		if _, err := sha.Write(buf[:n]); err != nil {
			return errutil.New("sha.Write", err)
		}
	}

	if skip {
		return nil
	}

	shasum := hex.EncodeToString(sha.Sum(nil))
	if strings.ToLower(d.SHA256) != shasum {
		return errutil.WithFrame(
			fmt.Errorf(
				"mismatch: %q (%q) expected %q",
				d.Filename,
				strings.ToLower(d.SHA256),
				shasum,
			),
		)
	}

	return nil
}

// ensureDownloadParams ensures the values of the download parameters are not empty.
func (d *Download) ensureDownloadParams() error {
	if d.URL == "" {
		return ErrDownloadURLEmpty
	}

	if d.Directory == "" {
		return ErrDownloadPathEmpty
	}

	if d.Filename == "" {
		return ErrDownloadNameEmpty
	}

	return nil
}
