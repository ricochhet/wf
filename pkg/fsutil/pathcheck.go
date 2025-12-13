package fsutil

import (
	"regexp"
	"slices"
	"strings"
)

type PathChecks []PathCheck

type PathCheck struct {
	Type   string
	Target string
	Action string
}

const (
	PathCheckTypeEndsWith  string = "EndsWith"
	PathCheckTypeContains  string = "Contains"
	PathCheckTypeDriveRoot string = "DriveRoot"

	PathCheckActionWarn string = "Warn"
	PathCheckActionDeny string = "Deny"
)

// CheckPathForProblemLocations checks if the file name ends with a given target.
func (pc *PathChecks) CheckPathForProblemLocations(path string) (bool, PathCheck) {
	path = strings.ToLower(Normalize(TrimPath(path)))
	parts := strings.Split(path, "/")
	check := PathCheck{}

	for _, check := range *pc {
		switch check.Type {
		case PathCheckTypeEndsWith:
			if strings.EqualFold(
				strings.ToLower(parts[len(parts)-1]),
				strings.ToLower(check.Target),
			) {
				return true, check
			}
		case PathCheckTypeContains:
			if slices.Contains(parts, strings.ToLower(check.Target)) {
				return true, check
			}
		case PathCheckTypeDriveRoot:
			return regexp.MustCompile(`^\w:(\\|\/)$`).Match([]byte(path)), check
		}
	}

	return false, check
}

// NewDefaultProblemPaths creates a default PathCheck slice.
func NewDefaultProblemPaths() []PathCheck {
	return []PathCheck{
		{Type: PathCheckTypeEndsWith, Target: "SteamApps", Action: PathCheckActionWarn},
		{Type: PathCheckTypeEndsWith, Target: "Documents", Action: PathCheckActionWarn},
		{Type: PathCheckTypeEndsWith, Target: "Desktop", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "Desktop", Action: PathCheckActionWarn},
		{Type: PathCheckTypeContains, Target: "scoped_dir", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "Downloads", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "OneDrive", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "NextCloud", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "DropBox", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "Google", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "Program Files", Action: PathCheckActionDeny},
		{Type: PathCheckTypeContains, Target: "Program Files (x86)", Action: PathCheckActionDeny},
		// {Type: PathCheckTypeContains, Target: "Windows", Action: PathCheckActionDeny},
		{Type: PathCheckTypeDriveRoot, Target: "", Action: PathCheckActionDeny},

		// Reserved words.
		{Type: PathCheckTypeEndsWith, Target: "CON", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "PRN", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "AUX", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "CLOCK$", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "NUL", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM0", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM1", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM2", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM3", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM4", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM5", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM6", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM7", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM8", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "COM9", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT0", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT1", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT2", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT3", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT4", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT5", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT6", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT7", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT8", Action: PathCheckActionDeny},
		{Type: PathCheckTypeEndsWith, Target: "LPT9", Action: PathCheckActionDeny},
	}
}
