package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

type MissionTagsFilter struct {
	Items map[string]struct{}
}

var missionTagsFilter = NewMissionTagsFilter()

// NewMissionTagsFilter creates a default MissionTagsFilter.
func NewMissionTagsFilter() *MissionTagsFilter {
	return &MissionTagsFilter{
		Items: map[string]struct{}{
			"PvpNode0":     {},
			"PvpNode9":     {},
			"PvpNode10":    {},
			"MercuryHUB":   {},
			"EarthHUB":     {},
			"TradeHUB1":    {},
			"SaturnHUB":    {},
			"EventNode763": {}, // The Index: Endurance (unused).
			"PlutoHUB":     {},
			"ZarimanHub":   {},
			"SolNode234":   {}, // Dormizone.
		},
	}
}

// NewMissionsCheat creates a new mission cheat instance.
func NewMissionsCheat(
	logger *logutil.Logger,
	regions, inventory []byte,
	index int,
) ([]byte, error) {
	missions, err := jsonutil.ResultAsArray(inventory, "Missions", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportRegions := gjson.ParseBytes(regions).Map()
	seen := make(map[string]struct{}, len(missions))
	combined := []string{}

	for _, item := range missions {
		seen[item.Get("Tag").String()] = struct{}{}
		logutil.Debugf(logger, "adding node from Missions: %s\n", item.Raw)
		combined = append(combined, item.Raw)
	}

	for uniqueName := range exportRegions {
		if _, exists := seen[uniqueName]; !exists {
			if _, blacklisted := missionTagsFilter.Items[uniqueName]; blacklisted {
				logutil.Infof(logger, "Skipping blacklisted Mission: %s\n", uniqueName)
				continue
			}

			node, err := database.NewNode(uniqueName, 1, 1)
			if err != nil {
				return nil, errutil.New("database.NewNode", err)
			}

			logutil.Debugf(logger, "adding node from ExportRegions: %s\n", node)
			combined = append(combined, node)
			seen[uniqueName] = struct{}{}
		}
	}

	result := []string{}

	for _, raw := range combined {
		tag := gjson.Get(raw, "Tag").String()
		if _, blacklisted := missionTagsFilter.Items[tag]; blacklisted {
			logutil.Infof(logger, "Skipping blacklisted Mission: %s\n", tag)
			continue
		}

		logutil.Debugf(logger, "adding node from combined: %s\n", raw)
		result = append(result, raw)
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "Missions", result, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(missions), len(result)-len(missions), len(result))

	return newInventory, nil
}
