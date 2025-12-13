package cmdutil

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/ricochhet/pkg/logutil"
)

// NewScanner creates a basic input scanner.
//
//	"fn" is called when enter is pressed.
func NewScanner(fn func(string) error) {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("> ")

	if !scanner.Scan() {
		logutil.Errorf(os.Stderr, "Failed to read input or input was empty.\n")
		return
	}

	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		logutil.Errorf(os.Stderr, "No command entered.\n")
		return
	}

	if err := fn(scanner.Text()); err != nil {
		logutil.Errorf(os.Stderr, "Command failed: %v\n", err)
	}

	pause()
}

// pause pauses the output so it can be visualized before closing.
func pause() {
	logutil.Infof(os.Stdout, "Press Enter to continue...\n")
	bufio.NewScanner(os.Stdin).Scan()
}
