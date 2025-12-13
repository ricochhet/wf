package internal

import (
	"sync"

	"github.com/ricochhet/dbmod/config"
	"github.com/ricochhet/dbmod/drivers"
	"github.com/ricochhet/pkg/pipeline"
)

type Database struct {
	Inventory []byte
	Stats     []byte
}

type Context struct {
	Mu       *sync.Mutex
	Flags    *config.Flags
	Conn     *drivers.MongoConnector
	Exports  *config.ExportManager
	Pipeline *pipeline.Pipeline[Database]
	Index    int
}
