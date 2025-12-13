package cheats

import (
	"fmt"

	"github.com/ricochhet/dbmod/internal/database"
	"github.com/ricochhet/pkg/errutil"
	"github.com/ricochhet/pkg/jsonutil"
	"github.com/ricochhet/pkg/logutil"
)

type AccoladesOptions struct {
	Staff     bool
	Founder   int64
	Guide     int64
	Moderator bool
	Partner   bool
	Heirloom  bool
	Counselor bool
}

// NewAccoladesCheat creates a new weapon skin cheat instance.
func (opt AccoladesOptions) NewAccoladesCheat(
	logger *logutil.Logger,
	inventory []byte,
	index int,
) ([]byte, error) {
	newAccolades, err := database.NewAccolades(opt.Heirloom)
	if err != nil {
		return nil, errutil.New("database.NewAccolades", err)
	}

	result, err := jsonutil.SetFieldInRawBytes(inventory, "Accolades", newAccolades, index)
	if err != nil {
		return nil, errutil.New("jsonutil.SetFieldInRawBytes", err)
	}

	accolades := map[string]any{
		"Staff":     opt.Staff,
		"Founder":   opt.Founder,
		"Guide":     opt.Guide,
		"Moderator": opt.Moderator,
		"Partner":   opt.Partner,
		"Counselor": opt.Counselor,
	}

	newInventory := string(result)

	for k, v := range accolades {
		logutil.Infof(logger, "Added: %s: %v\n", k, v)

		newInventory, err = jsonutil.SetFieldInBytes(newInventory, k, v, index)
		if err != nil {
			return nil, errutil.New(fmt.Sprintf("jsonutil.SetFieldInBytes (%s)", k), err)
		}
	}

	return []byte(newInventory), nil
}
