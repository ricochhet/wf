package internal

import (
	"github.com/ricochhet/dbmod/internal/patches"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/pipeline"
)

// NewInventoryPatches creates a new inventory patch instance.
func (ctx *Context) NewInventoryPatches() []*pipeline.Step[Database] {
	return []*pipeline.Step[Database]{
		{
			Name:    "xpInfo",
			Aliases: []string{},
			Step:    ctx.newXPInfoPatch,
		},
		{
			Name:    "shipDecorations",
			Aliases: []string{},
			Step:    ctx.newShipDecorationsPatch,
		},
	}
}

// newXPInfoPatch creates a new xp info patch instance.
func (ctx *Context) newXPInfoPatch(logger *logutil.Logger, db *Database) (*Database, error) {
	weapons := ctx.Exports.All().Weapons
	if len(weapons) == 0 {
		return nil, errutil.WithFramef("Weapons data is %d bytes", len(weapons))
	}

	warframes := ctx.Exports.All().Warframes
	if len(warframes) == 0 {
		return nil, errutil.WithFramef("Warframes data is %d bytes", len(warframes))
	}

	sentinels := ctx.Exports.All().Sentinels
	if len(sentinels) == 0 {
		return nil, errutil.WithFramef("Sentinels data is %d bytes", len(sentinels))
	}

	bytes, err := patches.NewXPInfoPatch(
		logger,
		weapons,
		warframes,
		sentinels,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newShipDecorationsPatch creates a new ship decoration patch instance.
func (ctx *Context) newShipDecorationsPatch(
	logger *logutil.Logger,
	db *Database,
) (*Database, error) {
	resources := ctx.Exports.All().Resources
	if len(resources) == 0 {
		return nil, errutil.WithFramef("Resources data is %d bytes", len(resources))
	}

	bytes, err := patches.NewShipDecorationsPatch(
		logger,
		resources,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}
