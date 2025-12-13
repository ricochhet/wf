package main

import (
	"github.com/ricochhet/dbmod/internal"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/sliceutil"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// NewInventoryCheats creates a new inventory cheat instance.
func NewInventoryCheats(names []string) error {
	ctx.Pipeline.Add(ctx.NewInventoryCheats()...)

	bytes, collection, err := readCollection(InventoriesCollection)
	if err != nil {
		return errutil.New("readCollection", err)
	}

	if err := writeBackup(InventoriesCollection, bytes); err != nil {
		return errutil.New("writeBackup", err)
	}

	err = commit(collection,
		ctx.Pipeline.Start(&internal.Database{Inventory: bytes, Stats: nil}, func(s []string) bool {
			return skip(s, names)
		}).Inventory,
		"inventories_commit")
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// NewStatsCheats creates a new stats cheat instance.
func NewStatsCheats(names []string) error {
	ctx.Pipeline.Add(ctx.NewStatsCheats()...)

	bytes, collection, err := readCollection(StatsCollection)
	if err != nil {
		return errutil.New("readCollection", err)
	}

	if err := writeBackup(StatsCollection, bytes); err != nil {
		return errutil.New("writeBackup", err)
	}

	err = commit(collection,
		ctx.Pipeline.Start(&internal.Database{Inventory: nil, Stats: bytes}, func(s []string) bool {
			return skip(s, names)
		}).Stats,
		"stats_commit")
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// NewInventoryPatches creates a new inventory patch instance.
func NewInventoryPatches(names []string) error {
	ctx.Pipeline.Add(ctx.NewInventoryPatches()...)

	bytes, collection, err := readCollection(InventoriesCollection)
	if err != nil {
		return errutil.New("readCollection", err)
	}

	if err := writeBackup(InventoriesCollection, bytes); err != nil {
		return errutil.New("writeBackup", err)
	}

	err = commit(collection,
		ctx.Pipeline.Start(&internal.Database{Inventory: bytes, Stats: nil}, func(s []string) bool {
			return skip(s, names)
		}).Inventory,
		"inventories_commit")
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// skip skips the given name if it's not contained the target slice.
func skip(names, target []string) bool {
	if sliceutil.Matches(names, target) != 0 || target[0] == "all" {
		return false
	}

	return true
}

// commit writes to the collection and creates the backup.
func commit(collection *mongo.Collection, data []byte, name string) error {
	if !ctx.Flags.DryRun {
		err := writeCollection(collection, data)
		if err != nil {
			return errutil.WithFrame(err)
		}
	}

	if err := writeBackup(name, data); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// readCollection reads from the database.
func readCollection(name string) ([]byte, *mongo.Collection, error) {
	mu.Lock()
	defer mu.Unlock()

	data, collection, err := ctx.Conn.Get(ctx.Flags.Database, name)
	if err != nil {
		return nil, nil, errutil.WithFrame(err)
	}

	return data, collection, nil
}

// writeCollection writes to the database.
func writeCollection(collection *mongo.Collection, data []byte) error {
	mu.Lock()
	defer mu.Unlock()

	if err := ctx.Conn.Set(collection, data); err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}
