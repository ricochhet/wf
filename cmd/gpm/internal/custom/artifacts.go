package custom

import (
	"context"
	"errors"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/avast/retry-go"
	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/arc"
	"github.com/ricochhet/pkg/cryptoutil"
	"github.com/ricochhet/pkg/dlutil"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
)

type ExtractionMode int

const (
	ExtractImmediately ExtractionMode = iota
	ExtractAfterAll
	ExtractNever
)

var (
	ErrNotArchive       = errors.New("not an archive")
	ErrAlreadyExtracted = errors.New("already extracted")
	ErrNeverExtracted   = errors.New("never extracted")
)

type archiveJob struct {
	Tarball string
	Dir     string
}

// messenger returns the download messenger to use.
func messenger(logger *logutil.Logger) *dlutil.Messenger {
	return &dlutil.Messenger{
		Start: func(f string) {
			logutil.Infof(logger, "Downloading: %s\n", f)
		},
	}
}

// retryOpts returns the []retry.Option to use.
func retryOpts(logger *logutil.Logger) *[]retry.Option {
	return &[]retry.Option{
		retry.Attempts(5),
		retry.LastErrorOnly(true),
		retry.OnRetry(func(n uint, err error) {
			logutil.Infof(logger, "Retry %d: %v\n", n, err)
		}),
	}
}

// Download downloads the specified data, and extracts based on the receiver value.
func (m ExtractionMode) Download(
	logger *logutil.Logger,
	downloads []config.Download,
	flags *config.Flags,
) error {
	dlutil.RetryOpts(*retryOpts(logger))

	var jobs []archiveJob

	for _, dl := range downloads {
		if len(dl.Platforms) != 0 && !slices.Contains(dl.Platforms, runtime.GOOS) {
			continue
		}

		if dl.Optional && !flags.Optionals {
			continue
		}

		job, err := m.download(logger, *messenger(logger), dl)

		switch {
		case errors.Is(err, ErrNotArchive),
			errors.Is(err, ErrAlreadyExtracted),
			errors.Is(err, ErrNeverExtracted):
		case err != nil:
			logutil.Infof(logger, "Failed to download %s: %v\n", dl.URL, err)
			return errutil.New("download", err)
		}

		if job != nil {
			jobs = append(jobs, *job)
		}
	}

	if m == ExtractAfterAll {
		for _, job := range jobs {
			if err := unarchive(logger, job.Tarball, job.Dir); err != nil {
				return errutil.New("unarchive", err)
			}
		}
	}

	return nil
}

// Remove attempts to remove all files within the files slice.
// If a SHA256 is specified, it will only remove if the hash matches.
func Remove(logger *logutil.Logger, files []config.File) error {
	for _, dl := range files {
		path, err := fsutil.FromCwd(dl.Name)
		if err != nil {
			return errutil.New("fsutil.FromCwd", err)
		}

		if !fsutil.Exists(path) {
			continue
		}

		if dl.Sha != "" {
			sum, err := cryptoutil.NewSHA256(path)
			if err != nil {
				return errutil.New("cryptoutil.NewSHA256", err)
			}

			if dl.Sha != sum {
				continue
			}
		}

		logutil.Infof(logger, "Removing: %s\n", path)

		if err := fsutil.RemoveAll(path); err != nil {
			logutil.Infof(logger, "Failed to remove file: %s\n", path)
			continue
		}
	}

	return nil
}

// download downloads the file and assigns an extraction job.
func (m ExtractionMode) download(
	logger *logutil.Logger,
	messenger dlutil.Messenger,
	dl config.Download,
) (*archiveJob, error) {
	arcs := []string{".7z", ".rar", ".zip", ".tar", ".gz"}

	filename, err := fsutil.URLFilename(dl.URL)
	if err != nil {
		return nil, errutil.New("fsutil.URLFilename", err)
	}

	if dl.Filename != "" {
		filename = dl.Filename
	}

	tarball := filepath.Join(dl.Dir, filename)
	d := dlutil.Download{
		URL:       dl.URL,
		Directory: dl.Dir,
		Filename:  filename,
		SHA256:    dl.Sha,
	}

	if err := d.Download(context.TODO(), messenger, d.NewDefaultValidator); err != nil {
		return nil, errutil.New("Download", err)
	}

	ext := filepath.Ext(filename)
	if !slices.Contains(arcs, ext) && !dl.Force {
		return nil, ErrNotArchive
	}

	extract := dl.Dir
	if dl.Extract != "" {
		extract = dl.Extract
	}

	dir := filepath.Join(extract, fsutil.Filename(filename))
	if fsutil.Exists(dir) {
		return nil, ErrAlreadyExtracted
	}

	switch m {
	case ExtractImmediately:
		if err := unarchive(logger, tarball, dir); err != nil {
			return nil, errutil.New("unarchive", err)
		}
	case ExtractAfterAll:
		return &archiveJob{Tarball: tarball, Dir: dir}, nil
	case ExtractNever:
	}

	return nil, ErrNeverExtracted
}

// unarchive unarchives a tarball to a destination.
func unarchive(logger *logutil.Logger, tarball, dst string) error {
	logutil.Infof(logger, "Extracting %s to %s\n", tarball, dst)

	if err := arc.Unarchive(tarball, dst); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}
