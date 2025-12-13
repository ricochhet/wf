package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

type CodexScanOptions struct {
	MaxScans int64
}

// NewCodexScansCheat creates a new codex scan cheat instance.
func (opt CodexScanOptions) NewCodexScansCheat(
	logger *logutil.Logger,
	custom, codex, enemies, stats []byte,
	index int,
) ([]byte, error) {
	scans, err := jsonutil.ResultAsArray(stats, "Scans", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportCustomScans := gjson.ParseBytes(custom).Array()
	exportCodex := gjson.ParseBytes(codex).Get("objects").Map()
	exportEnemies := gjson.ParseBytes(enemies).Get("avatars").Map()
	seen := make(map[string]struct{}, len(scans))
	combined := []string{}

	for _, scan := range scans {
		t := scan.Get("type").String()

		changed, err := database.NewScan(t, max(scan.Get("scans").Int(), opt.MaxScans))
		if err != nil {
			return nil, errutil.New("database.NewScan", err)
		}

		if scan.Raw != changed {
			logutil.Infof(logger, "Old: %v, New: %v\n", scan.Raw, changed)
		}

		logutil.Debugf(logger, "adding scan from Scans: %s\n", changed)
		combined = append(combined, changed)
		seen[t] = struct{}{}
	}

	for _, scan := range exportCustomScans {
		t := scan.String()
		if _, exists := seen[t]; !exists {
			scan, err := database.NewScan(t, opt.MaxScans)
			if err != nil {
				return nil, errutil.New("database.NewScan", err)
			}

			logutil.Debugf(logger, "adding scan from ExportCustomScans: %s\n", scan)
			combined = append(combined, scan)
			seen[t] = struct{}{}
		}
	}

	for t := range exportCodex {
		if _, exists := seen[t]; !exists {
			scan, err := database.NewScan(t, opt.MaxScans)
			if err != nil {
				return nil, errutil.New("database.NewScan", err)
			}

			logutil.Debugf(logger, "adding scan from ExportCodex: %s\n", scan)
			combined = append(combined, scan)
			seen[t] = struct{}{}
		}
	}

	for t := range exportEnemies {
		if _, exists := seen[t]; !exists {
			scan, err := database.NewScan(t, opt.MaxScans)
			if err != nil {
				return nil, errutil.New("database.NewScan", err)
			}

			logutil.Debugf(logger, "adding scan from ExportEnemies: %s\n", scan)
			combined = append(combined, scan)
			seen[t] = struct{}{}
		}
	}

	newStats, err := jsonutil.SetSliceInRawBytes(stats, "Scans", combined, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(scans), len(combined)-len(scans), len(combined))

	return newStats, nil
}
