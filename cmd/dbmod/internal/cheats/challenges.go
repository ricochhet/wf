package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

// NewChallengesCheat creates a new challenge cheat instance.
func NewChallengesCheat(
	logger *logutil.Logger,
	achievements, inventory []byte,
	index int,
) ([]byte, error) {
	challengeProgress, err := jsonutil.ResultAsArray(inventory, "ChallengeProgress", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportAchievements := gjson.ParseBytes(achievements).Map()
	seen := make(map[string]struct{}, len(challengeProgress))
	combined := []string{}

	for _, entry := range challengeProgress {
		seen[entry.Get("Name").String()] = struct{}{}
		logutil.Debugf(
			logger,
			"adding challenge progress from ChallengeProgress: %s\n",
			entry.Raw,
		)
		combined = append(combined, entry.Raw)
	}

	for uniqueName, item := range exportAchievements {
		if _, ok := seen[uniqueName]; !ok {
			requiredCount := item.Get("requiredCount").Int()
			if !item.Get("requiredCount").Exists() {
				requiredCount = 1
			}

			challenge, err := database.NewChallengeProgress(requiredCount, uniqueName)
			if err != nil {
				return nil, errutil.New("database.NewChallengeProgress", err)
			}

			logutil.Debugf(
				logger,
				"adding challenge progress from ExportAchievements: %s\n",
				challenge,
			)
			combined = append(combined, challenge)
		}
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(
		inventory,
		"ChallengeProgress",
		combined,
		index,
	)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(challengeProgress), len(combined)-len(challengeProgress), len(combined))

	return newInventory, nil
}
