//nolint:tagliatelle // wontfix
package database

import (
	"encoding/json"
	"fmt"

	"github.com/ricochhet/pkg/cryptoutil"
	"github.com/ricochhet/pkg/errutil"
)

type Accolades struct {
	Heirloom bool `json:"Heirloom"`
}

type Item struct {
	ItemType  string `json:"ItemType"`
	ItemCount int64  `json:"ItemCount,omitempty"`
}

type XPInfo struct {
	ItemType string `json:"ItemType"`
	XP       int64  `json:"XP"`
}

type Node struct {
	Tag       string `json:"Tag"`
	Completes int64  `json:"Completes,omitempty"`
	Tier      int64  `json:"Tier,omitempty"`
}

type Scan struct {
	Type  string `json:"type"`
	Scans int64  `json:"scans"`
}

type WeaponSkin struct {
	ID struct {
		Oid string `json:"$oid"`
	} `json:"_id"`
	ItemType string `json:"ItemType"`
}

type ChallengeProgress struct {
	Progress  int64    `json:"Progress"`
	Name      string   `json:"Name"`
	Completed []string `json:"Completed"`
}

type Enemy struct {
	Type       string `json:"type"`
	Kills      int64  `json:"kills"`
	Assists    int64  `json:"assists"`
	Headshots  int64  `json:"headshots"`
	Captures   int64  `json:"captures,omitempty"`
	Executions int64  `json:"executions,omitempty"`
	Deaths     int64  `json:"deaths,omitempty"`
}

// NewAccolade creates a string representation of an Accolade.
func NewAccolades(heirloom bool) (string, error) {
	obj := Accolades{Heirloom: heirloom}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewItems creates a string representation of an Item.
func NewItem(itemType string, itemCount int64) (string, error) {
	obj := Item{ItemType: itemType, ItemCount: itemCount}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewXPInfo creates a string representation of an XPInfo.
func NewXPInfo(itemType string, xp int64) (string, error) {
	obj := XPInfo{ItemType: itemType, XP: xp}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewItemType creates a string representation of an Item with ItemCount omitted.
func NewItemType(itemType string) (string, error) {
	obj := Item{ItemType: itemType}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewNode creates a string representation of a Node.
func NewNode(tag string, completes, tier int64) (string, error) {
	obj := Node{Tag: tag, Completes: completes, Tier: tier}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewScan creates a string representation of a Scan.
func NewScan(t string, scans int64) (string, error) {
	obj := Scan{Type: t, Scans: scans}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewNewWeaponSkinScan creates a string representation of a WeaponSkin.
func NewWeaponSkin(itemType string) (string, error) {
	ws := WeaponSkin{ItemType: itemType}
	ws.ID.Oid = fmt.Sprintf("cb70cb70cb70cb70%08x", cryptoutil.CatBreadHash(itemType))
	b, err := json.Marshal(ws)

	return string(b), errutil.WithFrame(err)
}

// NewChallengeProgress creates a string representation of a ChallengeProgress.
func NewChallengeProgress(progress int64, name string) (string, error) {
	obj := ChallengeProgress{Progress: progress, Name: name}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}

// NewEnemy creates a string representation of an Enemy.
func NewEnemy(
	itemType string,
	kills, assists, headshots, captures, executions, deaths int64,
) (string, error) {
	obj := Enemy{
		Type:       itemType,
		Kills:      kills,
		Assists:    assists,
		Headshots:  headshots,
		Captures:   captures,
		Executions: executions,
		Deaths:     deaths,
	}
	b, err := json.Marshal(obj)

	return string(b), errutil.WithFrame(err)
}
