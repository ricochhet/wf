package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/tidwall/gjson"
)

type FlavourItemsFilter struct {
	Items map[string]struct{}
}

var flavourItemsFilter = NewFlavourItemsFilter()

// NewFlavourItemsFilter creates a default FlavourItemsFilter.
func NewFlavourItemsFilter() *FlavourItemsFilter {
	return &FlavourItemsFilter{
		Items: map[string]struct{}{
			// Kavat/Kubrow color templates.
			"/Lotus/Types/Game/KubrowPet/Colors/KubrowPetColorKavatBase":      {}, // Executioner Grey.
			"/Lotus/Types/Game/KubrowPet/Colors/KubrowPetColorKavatSecondary": {}, // Hyacinth Blue.
			"/Lotus/Types/Game/KubrowPet/Colors/KubrowPetColorKavatTertiary":  {}, // Regor Green.
			// Glyphs with default data.
			// CheruuAx glyphs have duplicates but are not included.
			"/Lotus/Types/StoreItems/AvatarImages/FanChannel/AvatarImageDramakins":    {},
			"/Lotus/Types/StoreItems/AvatarImages/FanChannel/AvatarImageSenastra":     {},
			"/Lotus/Types/StoreItems/AvatarImages/FanChannel/AvatarImageDesRPG":       {},
			"/Lotus/Types/StoreItems/AvatarImages/FanChannel/AvatarImageKacchi":       {},
			"/Lotus/Types/StoreItems/AvatarImages/FanChannel/AvatarImageLovinDaTacos": {},
			"/Lotus/Types/StoreItems/AvatarImages/AvatarImageCreatorWgrates":          {},
			"/Lotus/Types/StoreItems/AvatarImages/ImageConquera2022B":                 {},
			"/Lotus/Types/StoreItems/AvatarImages/ImageConquera2022C":                 {},
			"/Lotus/Types/StoreItems/AvatarImages/ImageConquera2022D":                 {},
			// Color palettes.
			"/Lotus/Types/StoreItems/SuitCustomizations/ColourPickerItemD": {}, // Duplicate of "Storm" palette.
		},
	}
}

// NewFlavourItemsCheat creates a new flavor item cheat instance.
func NewFlavourItemsCheat(
	logger *logutil.Logger,
	flavor, inventory []byte,
	index int,
) ([]byte, error) {
	flavourItems, err := jsonutil.ResultAsArray(inventory, "FlavourItems", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	exportFlavour := gjson.ParseBytes(flavor).Map()
	seen := make(map[string]struct{}, len(flavourItems))
	combined := []string{}

	for _, item := range flavourItems {
		seen[item.Get("ItemType").String()] = struct{}{}
		logutil.Debugf(logger, "adding flavouritem from FlavourItems: %s\n", item.Raw)
		combined = append(combined, item.Raw)
	}

	for uniqueName, item := range exportFlavour {
		if item.Get("name").String() == "" {
			if _, exists := flavourItemsFilter.Items[uniqueName]; !exists {
				flavourItemsFilter.Items[uniqueName] = struct{}{}
			}

			continue
		}

		if _, blacklisted := flavourItemsFilter.Items[uniqueName]; blacklisted {
			continue
		}

		if _, exists := seen[uniqueName]; !exists {
			itemType, err := database.NewItemType(uniqueName)
			if err != nil {
				return nil, errutil.New("database.NewItemType", err)
			}

			logutil.Debugf(logger, "adding flavouritem from ExportFlavour: %s\n", itemType)
			combined = append(combined, itemType)
			seen[uniqueName] = struct{}{}
		}
	}

	result := []string{}

	for _, raw := range combined {
		itemType := gjson.Get(raw, "ItemType").String()
		if _, blacklisted := flavourItemsFilter.Items[itemType]; blacklisted {
			logutil.Infof(logger, "Skipping blacklisted FlavourItem: %s\n", itemType)
			continue
		}

		logutil.Debugf(logger, "adding flavouritem from combined: %s\n", raw)
		result = append(result, raw)
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "FlavourItems", result, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(flavourItems), len(result)-len(flavourItems), len(result))

	return newInventory, nil
}
