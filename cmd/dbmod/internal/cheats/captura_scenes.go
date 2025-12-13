package cheats

import (
	"github.com/ricochhet/dbmod/config"
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

// NewCapturaScenesCheat creates a new captura scene cheat instance.
func NewCapturaScenesCheat(
	logger *logutil.Logger,
	resources, virtuals, inventory []byte,
	index int,
) ([]byte, error) {
	miscItems, err := jsonutil.ResultAsArray(inventory, "MiscItems", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportResources := gjson.ParseBytes(resources).Map()
	exportVirtuals := gjson.ParseBytes(virtuals).Map()
	seen := make(map[string]struct{}, len(miscItems))
	combined := []string{}

	for _, item := range miscItems {
		seen[item.Get("ItemType").String()] = struct{}{}
		logutil.Debugf(logger, "adding captura scene from MiscItems: %s\n", item.Raw)
		combined = append(combined, item.Raw)
	}

	resourceParents := make(map[string]string)

	for name, node := range exportResources {
		parent := node.Get("parentName").String()
		if parent != "" {
			resourceParents[name] = parent
		}
	}

	for name, node := range exportVirtuals {
		if _, ok := resourceParents[name]; !ok {
			parent := node.Get("parentName").String()
			if parent != "" {
				resourceParents[name] = parent
			}
		}
	}

	for name := range exportResources {
		if config.ResourceInheritsFromMap(
			resourceParents,
			name,
			"/Lotus/Types/Items/MiscItems/PhotoboothTile",
		) {
			if _, exists := seen[name]; !exists {
				item, err := database.NewItem(name, 1)
				if err != nil {
					return nil, errutil.New("database.NewItem", err)
				}

				logutil.Debugf(
					logger,
					"adding captura scene from ExportResources: %s\n",
					item,
				)
				combined = append(combined, item)
				seen[name] = struct{}{}
			}
		}
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "MiscItems", combined, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(miscItems), len(combined)-len(miscItems), len(combined))

	return newInventory, nil
}
