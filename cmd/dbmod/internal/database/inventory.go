package database

import (
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/tidwall/gjson"
)

type InventoryItems struct {
	LongGuns     []gjson.Result
	Pistols      []gjson.Result
	Melee        []gjson.Result
	Hoverboards  []gjson.Result
	OperatorAmps []gjson.Result
	MoaPets      []gjson.Result
	KubrowPets   []gjson.Result
}

// NewInventoryItems creates InventoryItems with data provided by the byte slice.
func NewInventoryItems(inventory []byte, index int) (*InventoryItems, error) {
	longGuns, err := jsonutil.ResultAsArray(inventory, "LongGuns", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (LongGuns)", err)
	}

	pistols, err := jsonutil.ResultAsArray(inventory, "Pistols", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (Pistols)", err)
	}

	melee, err := jsonutil.ResultAsArray(inventory, "Melee", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (Melee)", err)
	}

	hoverboards, err := jsonutil.ResultAsArray(inventory, "Hoverboards", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (Hoverboards)", err)
	}

	operatorAmps, err := jsonutil.ResultAsArray(inventory, "OperatorAmps", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (OperatorAmps)", err)
	}

	moaPets, err := jsonutil.ResultAsArray(inventory, "MoaPets", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (MoaPets)", err)
	}

	kubrowPets, err := jsonutil.ResultAsArray(inventory, "KubrowPets", index)
	if err != nil {
		return nil, errutil.New("jsonutil.ResultAsArray (KubrowPets)", err)
	}

	return &InventoryItems{
		LongGuns:     longGuns,
		Pistols:      pistols,
		Melee:        melee,
		Hoverboards:  hoverboards,
		OperatorAmps: operatorAmps,
		MoaPets:      moaPets,
		KubrowPets:   kubrowPets,
	}, nil
}
