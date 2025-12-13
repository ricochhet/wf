package patches

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

// newShipDecorationsPatch creates a new ship decoration patch instance.
func NewShipDecorationsPatch(
	logger *logutil.Logger,
	resources, inventory []byte,
	index int,
) ([]byte, error) {
	decorations, err := jsonutil.ResultAsArray(inventory, "ShipDecorations", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportResources := gjson.ParseBytes(resources).Map()

	uniqueNames := make(map[string]string, len(exportResources))
	for uniqueName, item := range exportResources {
		decoStr := item.Get("deco").String()
		if decoStr != "" {
			uniqueNames[decoStr] = uniqueName
		}
	}

	blacklisted := make(map[string]struct{}, len(decorations))
	propDecorations := make(map[string]int64)

	for _, deco := range decorations {
		itemType := deco.Get("ItemType").String()
		count := deco.Get("ItemCount").Int()

		if uniqueName, found := uniqueNames[itemType]; found {
			blacklisted[itemType] = struct{}{}
			logutil.Infof(logger, "%d : %s\n", count, itemType)
			propDecorations[uniqueName] += count
		}
	}

	result := []string{}

	for _, deco := range decorations {
		itemType := deco.Get("ItemType").String()

		if _, skip := blacklisted[itemType]; skip {
			logutil.Infof(logger, "Skipping blacklisted ShipDecoration: %s\n", itemType)
			continue
		}

		logutil.Debugf(logger, "adding item from ShipDecorations: %s\n", deco.Raw)
		result = append(result, deco.Raw)
	}

	for uniqueName, count := range propDecorations {
		item, err := database.NewItem(uniqueName, count)
		if err != nil {
			return nil, errutil.New("database.NewItem", err)
		}

		logutil.Debugf(logger, "adding item from PropDecorations: %s\n", item)
		result = append(result, item)
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "ShipDecorations", result, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(decorations), len(result)-len(decorations), len(result))

	return newInventory, nil
}
