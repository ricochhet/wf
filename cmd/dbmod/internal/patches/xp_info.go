package patches

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/strutil"
	"github.com/tidwall/gjson"
)

type XPInfoFilter struct {
	Items            map[string]struct{}
	WhitelistedItems map[string]struct{}
	XPEarningParts   map[string]struct{}
	EndsWith         []string
}

var (
	xpInfoFilter             = NewXPInfoFilter()
	productCategoryBreakdown = false
)

// XPInfoFilter creates a default XPInfoFilter.
func NewXPInfoFilter() *XPInfoFilter {
	return &XPInfoFilter{
		Items: map[string]struct{}{}, // dynamically filled by the xpInfo function.
		WhitelistedItems: map[string]struct{}{
			"/Lotus/Powersuits/Khora/Kavat/KhoraKavatPowerSuit":      {},
			"/Lotus/Powersuits/Khora/Kavat/KhoraPrimeKavatPowerSuit": {},
		},
		XPEarningParts: map[string]struct{}{
			"LWPT_BLADE":       {},
			"LWPT_GUN_BARREL":  {},
			"LWPT_AMP_OCULUS":  {},
			"LWPT_MOA_HEAD":    {},
			"LWPT_ZANUKA_HEAD": {},
			"LWPT_HB_DECK":     {},
		},
		EndsWith: []string{
			"PetWeapon",
		},
	}
}

// newXPInfoPatch creates a new xp info patch instance.
func NewXPInfoPatch(
	logger *logutil.Logger,
	weapons, warframes, sentinels, inventory []byte,
	index int,
) ([]byte, error) {
	xpInfo, err := jsonutil.ResultAsArray(inventory, "XPInfo", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray", err)
	}

	inventoryItems, err := database.NewInventoryItems(inventory, index)
	if err != nil {
		return nil, errutil.New("database.NewInventoryItems", err)
	}

	exportWeapons := gjson.ParseBytes(weapons).Map()
	exportWarframes := gjson.ParseBytes(warframes).Map()
	exportSentinels := gjson.ParseBytes(sentinels).Map()

	seen := make(map[string]struct{}, len(xpInfo))
	combined := []string{}

	for _, item := range xpInfo {
		itemType := item.Get("ItemType").String()
		seen[itemType] = struct{}{}

		logutil.Debugf(logger, "adding xpinfo from XPInfo: %s\n", item.Raw)
		combined = append(combined, item.Raw)
	}

	for _, item := range exportWarframes {
		exaltedItems := item.Get("exalted").Array()
		for _, exaltedItem := range exaltedItems {
			itemType := exaltedItem.String()
			if _, whitelisted := xpInfoFilter.WhitelistedItems[itemType]; whitelisted {
				continue
			}

			xpInfoFilter.Items[itemType] = struct{}{}
		}
	}

	combined = append(combined, collectXpInfo(logger, seen, *inventoryItems, exportWeapons)...)

	result := []string{}
	seen = make(map[string]struct{}, len(combined))

	for _, raw := range combined {
		itemType := gjson.Get(raw, "ItemType").String()
		if _, blacklisted := xpInfoFilter.Items[itemType]; blacklisted {
			logutil.Infof(logger, "Skipping blacklisted XPInfo: %s\n", itemType)
			continue
		}

		if strutil.EndsWithAny(itemType, xpInfoFilter.EndsWith) {
			logutil.Infof(logger, "Skipping blacklisted XPInfo: %s\n", itemType)
			continue
		}

		if exportWeapons[itemType].Index == 0 &&
			exportWarframes[itemType].Index == 0 &&
			exportSentinels[itemType].Index == 0 {
			logutil.Infof(logger, "Unknown ItemType: %s\n", itemType)
		}

		if _, exists := seen[itemType]; exists {
			continue
		}

		seen[itemType] = struct{}{}

		if productCategoryBreakdown {
			breakdownProductCategories(
				logger,
				itemType,
				exportWeapons,
				exportWarframes,
				exportSentinels,
			)
		}

		logutil.Debugf(logger, "adding xpinfo from combined: %s\n", raw)
		result = append(result, raw)
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "XPInfo", result, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(xpInfo), len(result)-len(xpInfo), len(result))

	return newInventory, nil
}

// collectXpInfo collects all inventory item types and returns a string slice.
func collectXpInfo(
	logger *logutil.Logger,
	seen map[string]struct{},
	inventoryItems database.InventoryItems,
	weapons map[string]gjson.Result,
) []string {
	result := []string{}

	sets := [][]gjson.Result{
		inventoryItems.LongGuns,
		inventoryItems.Pistols,
		inventoryItems.Melee,
		inventoryItems.Hoverboards,
		inventoryItems.OperatorAmps,
		inventoryItems.MoaPets,
		inventoryItems.KubrowPets,
	}

	for _, set := range sets {
		result = append(result, xpInfo(logger, seen, set, weapons)...)
	}

	return result
}

// breakdownProductCategories prints a count of all items in each category.
func breakdownProductCategories(logger *logutil.Logger,
	itemType string, weapons, warframes, sentinels map[string]gjson.Result,
) {
	result := map[string][]string{}

	weapon := weapons[itemType].Get("productCategory").String()
	warframe := warframes[itemType].Get("productCategory").String()
	sentinel := sentinels[itemType].Get("productCategory").String()

	if weapon != "" {
		result[weapon] = append(result[weapon], itemType)
	}

	if warframe != "" {
		result[warframe] = append(result[warframe], itemType)
	}

	if sentinel != "" {
		result[sentinel] = append(result[sentinel], itemType)
	}

	for category, items := range result {
		logutil.Infof(logger, "%s (%d)\n", category, len(items))
	}
}

// xpInfo creates a string slice of parent ItemType objects from ModularParts.
func xpInfo(
	logger *logutil.Logger,
	seen map[string]struct{},
	items []gjson.Result,
	weapons map[string]gjson.Result,
) []string {
	result := []string{}

	for _, item := range items {
		xp := item.Get("XP").Int()
		itemType := item.Get("ItemType").String()

		for _, modularPart := range item.Get("ModularParts").Array() {
			uniqueName := modularPart.String()

			part := weapons[uniqueName]
			partType := part.Get("partType").String()

			if partType == "" {
				logutil.Debugf(logger, "part type was empty: %s\n", uniqueName)
				continue
			}

			if _, earnsXp := xpInfoFilter.XPEarningParts[partType]; !earnsXp {
				continue
			}

			xpInfoFilter.Items[itemType] = struct{}{}

			if _, exists := seen[uniqueName]; exists {
				continue
			}

			xpInfo, err := database.NewXPInfo(uniqueName, xp)
			if err != nil {
				return nil
			}

			logutil.Debugf(logger, "adding xpinfo from ModularPart: %s\n", xpInfo)

			result = append(result, xpInfo)
		}
	}

	return result
}
