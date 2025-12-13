package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

type EnemyStatOptions struct {
	Kills     int64
	Assists   int64
	Headshots int64
}

// NewEnemyStatsCheat creates a new enemy stat cheat instance.
func (opt EnemyStatOptions) NewEnemyStatsCheat(
	logger *logutil.Logger,
	enemies, stats []byte,
	index int,
) ([]byte, error) {
	enemyStats, err := jsonutil.ResultAsArray(stats, "Enemies", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportEnemies := gjson.ParseBytes(enemies).Get("avatars").Map()
	seen := make(map[string]struct{}, len(enemyStats))
	combined := []string{}

	for _, enemy := range enemyStats {
		itemType := enemy.Get("type").String()
		seen[itemType] = struct{}{}

		changed, err := database.NewEnemy(itemType,
			max(enemy.Get("kills").Int(), opt.Kills),
			max(enemy.Get("assists").Int(), opt.Assists),
			max(enemy.Get("headshots").Int(), opt.Headshots),
			max(0, enemy.Get("captures").Int()),
			max(0, enemy.Get("executions").Int()),
			max(0, enemy.Get("deaths").Int()))
		if err != nil {
			return nil, errutil.New("database.NewEnemy", err)
		}

		if enemy.Raw != changed {
			logutil.Infof(logger, "Old: %v, New: %v\n", enemy.Raw, changed)
		}

		logutil.Debugf(logger, "adding enemy from Enemies: %s\n", changed)
		combined = append(combined, changed)
	}

	for uniqueName := range exportEnemies {
		if _, exists := seen[uniqueName]; !exists {
			enemy, err := database.NewEnemy(uniqueName,
				opt.Kills,
				opt.Assists,
				opt.Headshots, 0, 0, 0)
			if err != nil {
				return nil, errutil.New("database.NewEnemy", err)
			}

			logutil.Debugf(logger, "adding enemy from ExportEnemies: %s\n", enemy)
			combined = append(combined, enemy)
			seen[uniqueName] = struct{}{}
		}
	}

	newStats, err := jsonutil.SetSliceInRawBytes(stats, "Enemies", combined, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(enemyStats), len(combined)-len(enemyStats), len(combined))

	return newStats, nil
}
