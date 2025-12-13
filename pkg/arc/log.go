package arc

import (
	"log"
	"os"
)

var DEBUG = os.Getenv("ARC_DEBUG") == "true"

func logf(format string, a ...any) {
	if DEBUG {
		log.Printf(format, a...)
	}
}
