package cheats

import (
	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/strutil"
	"github.com/tidwall/gjson"
)

type WeaponSkinFilter struct {
	ItemTypes               map[string]struct{}
	EndsWith                []string
	Icons                   map[string]struct{}
	Descriptions            map[string]struct{}
	WhitelistedDescriptions map[string]struct{}
}

var weaponSkinFilter = NewWeaponSkinFilter()

// NewWeaponSkinFilter creates a default WeaponSkinFilter.
func NewWeaponSkinFilter() *WeaponSkinFilter {
	return &WeaponSkinFilter{
		ItemTypes: map[string]struct{}{
			// Internal (base) items.
			"/Lotus/Upgrades/Skins/Effects/BaseFootsteps":                    {},
			"/Lotus/Upgrades/Skins/Operator/AnimationSets/BaseOperatorAnims": {},
			// Unused auxiliary cosmetics.
			"/Lotus/Upgrades/Skins/Halos/PrototypeRaidHalo": {},
			// Unreleased / unfinished cosmetics.
			"/Lotus/Upgrades/Skins/Operator/Hair/HairAdultNightwave":  {},
			"/Lotus/Upgrades/Skins/Operator/Hair/HairAdultNightwaveB": {},
			"/Lotus/Upgrades/Skins/Promo/ChangYou/CYSingleStaffSkin":  {},
			// "/Lotus/Upgrades/Skins/Weapons/UnrealTournament/DrakgoonFlakCannonSkinPrimaryProjectileSkin"
			// "/Lotus/Upgrades/Skins/Weapons/UnrealTournament/OgrisRocketLauncherSkinPrimaryProjectileSkin"
			// "/Lotus/Upgrades/Skins/Weapons/UnrealTournament/StahltaShockRifleSkinPrimaryProjectileSkin"
			// "/Lotus/Upgrades/Skins/Weapons/UnrealTournament/StahltaShockRifleSkinSecondaryProjectileSkin"
			// Default (base) cosmetics.
			"/Lotus/Types/Game/InfestedKavatPet/Patterns/InfestedCritterPatternDefault":     {},
			"/Lotus/Types/Game/InfestedPredatorPet/Patterns/InfestedPredatorPatternDefault": {},
			"/Lotus/Upgrades/Skins/Excalibur/ExcaliburPrimeAlabasterSkin":                   {},
			"/Lotus/Upgrades/Skins/Saryn/WF1999SarynHelmet":                                 {},
		},
		EndsWith: []string{
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/InfNightWaveWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/InfestedWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/ColtekWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/DiamondWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/DomeWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/IctusPrimeWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/IctusWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/JetWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/PrismaJetWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/KavatPetWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/DethcubePrimeWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/ParrotWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/OrokinWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/PrimeSentinelWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/GardenerWingsRight"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/GardenerWingsStatic"
			// "/Lotus/Upgrades/Skins/Sentinels/Wings/GardenerWingsStaticRight"
			"WingsRight", "WingsStaticRight",
		},
		Icons: map[string]struct{}{
			"/Lotus/Interface/Icons/StoreIcons/Resources/CraftingComponents/GenericWarframeHelmet.png": {},
		},
		Descriptions: map[string]struct{}{
			"/Lotus/Language/Items/GenericSuitCustomizationDesc": {},
		},
		WhitelistedDescriptions: map[string]struct{}{
			"/Lotus/Language/Items/GenericOperatorHairDescription": {},
			"/Lotus/Language/Operator/DrifterBeardDesc":            {},
		},
	}
}

// NewWeaponSkinsCheat creates a new weapon skin cheat instance.
func NewWeaponSkinsCheat(
	logger *logutil.Logger,
	customs, inventory []byte,
	index int,
) ([]byte, error) {
	skins, err := jsonutil.ResultAsArray(inventory, "WeaponSkins", index)
	if err != nil {
		return nil, errutil.New("jsonutil.Result", err)
	}

	exportCustoms := gjson.ParseBytes(customs).Map()
	seen := make(map[string]struct{}, len(skins))
	combined := []string{}

	for _, skin := range skins {
		itemType := skin.Get("ItemType").String()
		if _, blacklisted := weaponSkinFilter.ItemTypes[itemType]; blacklisted {
			logutil.Infof(logger, "Skipping blacklisted WeaponSkin: %s\n", itemType)
			continue
		}

		seen[itemType] = struct{}{}

		logutil.Debugf(logger, "adding weaponskin from WeaponSkins: %s\n", skin.Raw)
		combined = append(combined, skin.Raw)
	}

	for uniqueName, item := range exportCustoms {
		if item.Get("name").String() == "" {
			if _, whitelisted := weaponSkinFilter.WhitelistedDescriptions[item.Get("description").String()]; !whitelisted {
				weaponSkinFilter.ItemTypes[uniqueName] = struct{}{}
				continue
			}
		}

		if _, blacklisted := weaponSkinFilter.Descriptions[item.Get("description").String()]; blacklisted {
			weaponSkinFilter.ItemTypes[uniqueName] = struct{}{}
			continue
		}

		if _, blacklisted := weaponSkinFilter.Icons[item.Get("icon").String()]; blacklisted {
			weaponSkinFilter.ItemTypes[uniqueName] = struct{}{}
			continue
		}

		if strutil.EndsWithAny(uniqueName, weaponSkinFilter.EndsWith) {
			weaponSkinFilter.ItemTypes[uniqueName] = struct{}{}
			continue
		}

		if _, exists := seen[uniqueName]; exists {
			continue
		}

		seen[uniqueName] = struct{}{}

		weaponSkin, err := database.NewWeaponSkin(uniqueName)
		if err != nil {
			return nil, errutil.New("database.NewWeaponSkin", err)
		}

		logutil.Debugf(logger, "adding weaponskin from ExportCustoms: %s\n", weaponSkin)
		combined = append(combined, weaponSkin)
	}

	result := []string{}

	for _, raw := range combined {
		itemType := gjson.Get(raw, "ItemType").String()
		if _, blacklisted := weaponSkinFilter.ItemTypes[itemType]; blacklisted {
			logutil.Infof(logger, "Skipping blacklisted WeaponSkin: %s\n", itemType)
			continue
		}

		logutil.Debugf(logger, "adding weaponskin from combined: %s\n", raw)
		result = append(result, raw)
	}

	newInventory, err := jsonutil.SetSliceInRawBytes(inventory, "WeaponSkins", result, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetSliceInRawBytes", err)
	}

	logutil.Infof(logger, "Original: %d, Added: %d, Final: %d\n",
		len(skins), len(result)-len(skins), len(result))

	return newInventory, nil
}
