package cueutil

import (
	"os"

	"github.com/ricochhet/pkg/timeutil"
)

type Builtins[T any] struct {
	Args             []T
	WorkingDirectory string
	Time             string
}

// Map returns a map of the provided Builtins receiver.
func (b *Builtins[T]) Map() *map[string]any {
	return &map[string]any{
		"builtin": map[string]any{
			"args": b.Args,
			"cwd":  b.WorkingDirectory,
			"time": b.Time,
		},
	}
}

// NewBuiltins returns a new Builtins.
func NewBuiltins[T any](args []T) *Builtins[T] {
	wd, _ := os.Getwd()

	return &Builtins[T]{
		Args:             args,
		WorkingDirectory: wd,
		Time:             timeutil.NewDefaultTimestamp(),
	}
}
