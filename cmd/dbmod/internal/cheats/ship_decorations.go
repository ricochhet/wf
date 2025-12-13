package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

type ShipDecorationOptions struct {
	MaxShipDecorations int64
}

// NewShipDecorationsCheat creates a new ship decoration cheat instance.
func (opt ShipDecorationOptions) NewShipDecorationsCheat(
	logger *logutil.Logger,
	resources, inventory []byte,
	index int,
) ([]byte, error) {
	decorations, err := jsonutil.ResultAsArray(inventory, "ShipDecorations", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportResources := gjson.ParseBytes(resources).Map()
	seen := make(map[string]struct{}, len(decorations))
	combined := []string{}

	for _, deco := range decorations {
		itemType := deco.Get("ItemType").String()
		seen[itemType] = struct{}{}

		changed, err := database.NewItem(
			itemType,
			max(deco.Get("ItemCount").Int(), opt.MaxShipDecorations),
		)
		if err != nil {
			return nil, errutil.New("database.NewItem", err)
		}

		if deco.Raw != changed {
			logutil.Infof(logger, "Old: %v, New: %v\n", deco.Raw, changed)
		}

		logutil.Debugf(logger, "adding decoration from ShipDecorations: %s\n", changed)
		combined = append(combined, changed)
	}

	for uniqueName, item := range exportResources {
		if item.Get("productCategory").String() == "ShipDecorations" {
			if _, exists := seen[uniqueName]; !exists {
				item, err := database.NewItem(uniqueName, opt.MaxShipDecorations)
				if err != nil {
					return nil, errutil.New("database.NewItem", err)
				}

				logutil.Debugf(logger, "adding decoration ExportResources: %s\n", item)
				combined = append(combined, item)
				seen[uniqueName] = struct{}{}
			}
		}
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "ShipDecorations", combined, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(decorations), len(combined)-len(decorations), len(combined))

	return newInventory, nil
}
