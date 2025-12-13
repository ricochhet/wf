package logutil

import (
	"fmt"
	"io"
	"sync/atomic"
)

var debug atomic.Bool

// SetDebug sets the debug logging to the specified value.
func SetDebug(v bool) {
	debug.Store(v)
}

// IsDebug gets the debug state.
func IsDebug() bool {
	return debug.Load()
}

// Debugf prints if IsDebug is true.
func Debugf(w io.Writer, format string, a ...any) {
	if debug.Load() {
		fmt.Fprintf(w, "[debug] "+format, a...)
	}
}

// Warnf prints a warn log.
func Warnf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, "[warn] "+format, a...)
}

// Errorf prints a Error log.
func Errorf(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, "[error] "+format, a...)
}

// Infof prints a Info log.
func Infof(w io.Writer, format string, a ...any) {
	fmt.Fprintf(w, "[info] "+format, a...)
}
