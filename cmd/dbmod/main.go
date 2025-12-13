package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/ricochhet/dbmod/config"
	"github.com/ricochhet/dbmod/drivers"
	"github.com/ricochhet/dbmod/internal"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/fsutil"
	"github.com/ricochhet/pkg/logutil"
	"github.com/ricochhet/pkg/pipeline"
	"github.com/ricochhet/pkg/timeutil"
)

// version is the git tag at the time of build and is used to denote the
// binary's current version. This value is supplied as an ldflag at compile
// time by goreleaser (see .goreleaser.yml).
const (
	name     = "dbmod"
	version  = "0.1.0"
	revision = "HEAD"
)

func usage() {
	fmt.Fprint(os.Stderr, `Tasks:
  dbmod help [TASK]               # Show this help
  dbmod view                      # View all embedded files
  dbmod dump FILE                 # Dump an embedded file to disk
  dbmod [MODE] inventory COMMAND  # Run a specific cheat or patch
                                        challenges      (cheat)
                                        capturaScenes   (cheat)
                                        flavourItems    (cheat)
                                        missions        (cheat)
                                        shipDecorations (cheat, patch)
                                        weaponSkins     (cheat)
                                        xpInfo          (patch)
  dbmod [MODE] stats COMMAND     # Run a specific cheat or patch
                                       codexScans      (cheat)
                                       enemyStats      (cheat)
  dbmod version                  # Display dbmod version

Options:
`)
	flag.PrintDefaults()

	os.Exit(0)
}

var mu sync.Mutex

const (
	InventoriesCollection = "inventories"
	StatsCollection       = "stats"
)

var ctx internal.Context

// showVersion shows the current version of dbmod.
func showVersion() {
	logutil.Infof(os.Stdout, "%s\n", version)
	os.Exit(0)
}

// readExports reads all export data into bytes.
// prioritizes reading from a local file, falling back to embedded data.
func readExports(path string) {
	mu.Lock()
	defer mu.Unlock()

	ctx.Exports.SetAll(&config.Exports{
		Achievements: maybeRead(path + "/ExportAchievements.json"),
		Codex:        maybeRead(path + "/ExportCodex.json"),
		Customs:      maybeRead(path + "/ExportCustoms.json"),
		Enemies:      maybeRead(path + "/ExportEnemies.json"),
		Flavor:       maybeRead(path + "/ExportFlavour.json"),
		Regions:      maybeRead(path + "/ExportRegions.json"),
		Resources:    maybeRead(path + "/ExportResources.json"),
		Virtuals:     maybeRead(path + "/ExportVirtuals.json"),
		Weapons:      maybeRead(path + "/ExportWeapons.json"),
		Warframes:    maybeRead(path + "/ExportWarframes.json"),
		Sentinels:    maybeRead(path + "/ExportSentinels.json"),
		AllScans:     maybeRead(path + "/allScans.json"),
	})
}

// maybeRead reads a file from the specified path name.
func maybeRead(name string) []byte {
	data, err := fsutil.ReadBytes(name)
	if err != nil {
		logutil.Errorf(os.Stderr, "Failed to read data: %v\n", err)
		return nil
	}

	return data
}

func main() {
	var err error

	cfg := readConfig()

	ctx = internal.Context{
		Mu:       &mu,
		Flags:    cfg,
		Exports:  config.NewExportManager(),
		Pipeline: pipeline.NewPipeline[internal.Database](),
		Index:    cfg.Index,
	}

	readExports(cfg.WFData)

	dbCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ctx.Conn, err = drivers.NewMongoConnector(dbCtx, ctx.Flags.MongoURI)
	exitOnErr(err)

	defer func() {
		err := ctx.Conn.Disconnect()
		exitOnErr(err)
	}()

	logutil.LogTime.Store(ctx.Flags.LogTime)

	if flag.NArg() == 0 {
		exitOnErr(errutil.WithFrame(errors.New("no args specified")))
	}

	exitOnErr(commands())
}

// commands runs functions based on the provided Flags.Args[0].
func commands() error {
	var err error

	cmd := ctx.Flags.Args[0]
	subcmd := []string{"all"}

	if len(ctx.Flags.Args) > 1 {
		subcmd = ctx.Flags.Args[1:]
	}

	switch cmd {
	case "help":
		usage()
	case "i", "inventory":
		switch ctx.Flags.Mode {
		case "c", "cheat":
			if len(subcmd) >= 1 {
				err = NewInventoryCheats(subcmd)
			} else {
				usage()
			}
		case "p", "patch":
			if len(subcmd) >= 1 {
				err = NewInventoryPatches(subcmd)
			} else {
				usage()
			}
		default:
			usage()
		}
	case "s", "stats":
		switch ctx.Flags.Mode {
		case "c", "cheat":
			if len(subcmd) >= 1 {
				err = NewStatsCheats(subcmd)
			} else {
				usage()
			}
		default:
			usage()
		}
	case "version":
		showVersion()
	default:
		usage()
	}

	return err
}

// writeBackup writes a backup file to the Flag.DBData path.
func writeBackup(name string, data []byte) error {
	mu.Lock()
	defer mu.Unlock()

	if len(data) == 0 {
		return errors.New("size of data: 0")
	}

	err := fsutil.WriteBytes(
		filepath.Join(
			ctx.Flags.DBData,
			fmt.Sprintf("%s-%s", name, timeutil.NewDefaultTimestamp()),
		),
		data,
		0o644,
	)
	if err != nil {
		return errutil.WithFrame(err)
	}

	return nil
}

// exitOnErr prints an error message and exits the program.
func exitOnErr(err error) {
	if err != nil {
		logutil.Errorf(os.Stderr, "%s: %v\n", os.Args[0], err.Error())
		os.Exit(1)
	}
}
