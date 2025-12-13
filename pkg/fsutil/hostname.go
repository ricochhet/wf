package fsutil

import (
	"regexp"
	"strings"
)

var ReservedHostnames = []string{
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
	"PRN", "AUX", "NUL",
}

// IsValidHostname return true if the hostname is valid and does not contain reserved hostnames.
func IsValidHostname(hostname string) bool {
	if len(hostname) < 1 || len(hostname) > 15 {
		return false
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9-]+$`)
	if !re.MatchString(hostname) {
		return false
	}

	for _, reserved := range ReservedHostnames {
		if strings.EqualFold(hostname, reserved) {
			return false
		}
	}

	return true
}
