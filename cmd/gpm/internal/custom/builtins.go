package custom

import (
	"sync"

	"github.com/ricochhet/gpm/config"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
)

type Builtins struct {
	// Builtins
	Download string
	Remove   string
	// Internal
	mu        sync.Mutex
	artifacts config.Artifacts
}

// NewDefaultBuiltins returns a default Builtins struct.
func NewDefaultBuiltins() *Builtins {
	return &Builtins{
		Download: "gpm:pull",
		Remove:   "gpm:prune",
	}
}

// SetArtifacts sets the artifacts.
func (c *Builtins) SetArtifacts(a config.Artifacts) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.artifacts = a
}

// Start checks if the name matches a known function.
// If a function is known, execute, and return true. Otherwise return false.
func (c *Builtins) Start(
	logger *logutil.Logger,
	name string,
	flags config.Flags,
) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.ensureBuiltins(); err != nil {
		return true, errutil.WithFrame(err)
	}

	switch name {
	case c.Download:
		return true, errutil.WithFrame(
			ExtractImmediately.Download(logger, c.artifacts.Pull, &flags))
	case c.Remove:
		return true, errutil.WithFrame(Remove(logger, c.artifacts.Prune))
	default:
		return false, nil
	}
}

// ensureBuiltins ensures the builtin command values are not empty.
func (c *Builtins) ensureBuiltins() error {
	if c.Download == "" {
		return errutil.WithFramef("value of c.Download is empty")
	}

	if c.Remove == "" {
		return errutil.WithFramef("value of c.Remove is empty")
	}

	return nil
}
