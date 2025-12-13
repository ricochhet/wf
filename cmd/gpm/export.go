package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ricochhet/pkg/errutil"
)

// exportUpstart exports the procfile in upstart format.
func exportUpstart(path string) error {
	for i, proc := range ctx.SharedProc.All() {
		f, err := os.Create(filepath.Join(path, "app-"+proc.Name+".conf"))
		if err != nil {
			return errutil.New("os.Create", err)
		}

		fmt.Fprintf(f, "start on starting app-%s\n", proc.Name)
		fmt.Fprintf(f, "stop on stopping app-%s\n", proc.Name)
		fmt.Fprintf(f, "respawn\n")
		fmt.Fprintf(f, "\n")

		env := map[string]string{}

		taskfile, err := filepath.Abs(ctx.Flags.Taskfile)
		if err != nil {
			return errutil.New("filepath.Abs", err)
		}

		b, err := os.ReadFile(filepath.Join(filepath.Dir(taskfile), ".env"))
		if err == nil {
			for line := range strings.SplitSeq(string(b), "\n") {
				token := strings.SplitN(line, "=", 2)
				if len(token) != 2 {
					continue
				}

				token[0] = strings.TrimPrefix(token[0], "export ")
				token[0] = strings.TrimSpace(token[0])
				token[1] = strings.TrimSpace(token[1])
				env[token[0]] = token[1]
			}
		}

		if err != nil {
			return errutil.New("os.ReadFile", err)
		}

		fmt.Fprintf(f, "env PORT=%d\n", ctx.Flags.BasePort+uint(i))

		for k, v := range env {
			fmt.Fprintf(f, "env %s='%s'\n", k, strings.ReplaceAll(v, "'", "\\'"))
		}

		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "setuid app\n")
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "chdir %s\n", filepath.ToSlash(filepath.Dir(taskfile)))
		fmt.Fprintf(f, "\n")
		fmt.Fprintf(f, "exec %s\n", proc.Cmdline)

		f.Close()
	}

	return nil
}

// command: export.
func export(format, path string) error {
	err := os.MkdirAll(path, 0o755)
	if err != nil {
		return errutil.WithFrame(err)
	}

	if format == "upstart" {
		return exportUpstart(path)
	}

	return nil
}
