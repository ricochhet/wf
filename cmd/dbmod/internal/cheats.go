package internal

import (
	"github.com/ricochhet/dbmod/internal/cheats"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/pipeline"
)

// NewInventoryCheats creates a new inventory cheat instance.
func (ctx *Context) NewInventoryCheats() []*pipeline.Step[Database] {
	return []*pipeline.Step[Database]{
		{
			Name:    "accolades",
			Aliases: []string{},
			Step:    ctx.NewAccoladesCheat,
		},
		{
			Name:    "challenges",
			Aliases: []string{},
			Step:    ctx.newChallengesCheat,
		},
		{
			Name:    "capturaScenes",
			Aliases: []string{},
			Step:    ctx.newCapturaScenesCheat,
		},
		{
			Name:    "flavourItems",
			Aliases: []string{},
			Step:    ctx.newFlavourItemsCheat,
		},
		{
			Name:    "missions",
			Aliases: []string{},
			Step:    ctx.newMissionsCheat,
		},
		{
			Name:    "shipDecorations",
			Aliases: []string{},
			Step:    ctx.newShipDecorationsCheat,
		},
		{
			Name:    "weaponSkins",
			Aliases: []string{},
			Step:    ctx.newWeaponSkinsCheat,
		},
	}
}

// NewStatsCheats creates a new stats cheat instance.
func (ctx *Context) NewStatsCheats() []*pipeline.Step[Database] {
	return []*pipeline.Step[Database]{
		{
			Name:    "codexScans",
			Aliases: []string{},
			Step:    ctx.newCodexScansCheat,
		},
		{
			Name:    "enemyStats",
			Aliases: []string{},
			Step:    ctx.newEnemyStatsCheat,
		},
	}
}

// NewAccoladesCheat creates a new challenge cheat instance.
func (ctx *Context) NewAccoladesCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	opts := cheats.AccoladesOptions{
		Staff:     true,
		Founder:   4,
		Guide:     2,
		Moderator: true,
		Partner:   true,
		Heirloom:  true,
		Counselor: true,
	}

	bytes, err := opts.NewAccoladesCheat(logger, db.Inventory, ctx.Index)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newChallengesCheat creates a new challenge cheat instance.
func (ctx *Context) newChallengesCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	achievements := ctx.Exports.All().Achievements
	if len(achievements) == 0 {
		return nil, errutil.WithFramef("Achievements data is %d bytes", len(achievements))
	}

	bytes, err := cheats.NewChallengesCheat(
		logger,
		achievements,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newCapturaScenesCheat creates a new captura scene cheat instance.
func (ctx *Context) newCapturaScenesCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	resources := ctx.Exports.All().Resources
	if len(resources) == 0 {
		return nil, errutil.WithFramef("Resources data is %d bytes", len(resources))
	}

	virtuals := ctx.Exports.All().Virtuals
	if len(virtuals) == 0 {
		return nil, errutil.WithFramef("Virtuals data is %d bytes", len(virtuals))
	}

	bytes, err := cheats.NewCapturaScenesCheat(
		logger,
		resources,
		virtuals,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newFlavourItemsCheat creates a new flavor item cheat instance.
func (ctx *Context) newFlavourItemsCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	flavor := ctx.Exports.All().Flavor
	if len(flavor) == 0 {
		return nil, errutil.WithFramef("Flavor data is %d bytes", len(flavor))
	}

	bytes, err := cheats.NewFlavourItemsCheat(
		logger,
		flavor,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newMissionsCheat creates a new mission cheat instance.
func (ctx *Context) newMissionsCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	regions := ctx.Exports.All().Regions
	if len(regions) == 0 {
		return nil, errutil.WithFramef("Regions data is %d bytes", len(regions))
	}

	bytes, err := cheats.NewMissionsCheat(
		logger,
		regions,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newShipDecorationsCheat creates a new ship decoration cheat instance.
func (ctx *Context) newShipDecorationsCheat(
	logger *logutil.Logger,
	db *Database,
) (*Database, error) {
	opts := cheats.ShipDecorationOptions{MaxShipDecorations: 999}

	resources := ctx.Exports.All().Resources
	if len(resources) == 0 {
		return nil, errutil.WithFramef("Resources data is %d bytes", len(resources))
	}

	bytes, err := opts.NewShipDecorationsCheat(
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

// newWeaponSkinsCheat creates a new weapon skin cheat instance.
func (ctx *Context) newWeaponSkinsCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	customs := ctx.Exports.All().Customs
	if len(customs) == 0 {
		return nil, errutil.WithFramef("Customs data is %d bytes", len(customs))
	}

	bytes, err := cheats.NewWeaponSkinsCheat(
		logger,
		customs,
		db.Inventory,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: bytes, Stats: db.Stats}, nil
}

// newEnemyStatsCheat creates a new enemy stat cheat instance.
func (ctx *Context) newEnemyStatsCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	opts := cheats.EnemyStatOptions{Kills: 25, Assists: 5, Headshots: 10}

	enemies := ctx.Exports.All().Enemies
	if len(enemies) == 0 {
		return nil, errutil.WithFramef("Enemies data is %d bytes", len(enemies))
	}

	bytes, err := opts.NewEnemyStatsCheat(logger,
		enemies, db.Stats, ctx.Index)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: db.Inventory, Stats: bytes}, nil
}

// newCodexScansCheat creates a new codex scan cheat instance.
func (ctx *Context) newCodexScansCheat(logger *logutil.Logger, db *Database) (*Database, error) {
	opts := cheats.CodexScanOptions{MaxScans: 99}

	allScans := ctx.Exports.All().AllScans
	if len(allScans) == 0 {
		return nil, errutil.WithFramef("AllScans data is %d bytes", len(allScans))
	}

	codex := ctx.Exports.All().Codex
	if len(codex) == 0 {
		return nil, errutil.WithFramef("Codex data is %d bytes", len(codex))
	}

	enemies := ctx.Exports.All().Enemies
	if len(enemies) == 0 {
		return nil, errutil.WithFramef("Enemies data is %d bytes", len(enemies))
	}

	bytes, err := opts.NewCodexScansCheat(
		logger,
		allScans,
		codex,
		enemies,
		db.Stats,
		ctx.Index,
	)
	if err != nil {
		return nil, errutil.WithFrame(err)
	}

	return &Database{Inventory: db.Inventory, Stats: bytes}, nil
}
