package normalize

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"
)

// Debug is true when the SWAGGER_DEBUG env var is not empty.
//
// It enables a more verbose logging of this package.
var Debug = os.Getenv("SWAGGER_DEBUG") != ""

var (
	// normLogger is a debug logger for this package
	normLogger *log.Logger
)

func init() {
	debugOptions()
}

func debugOptions() {
	normLogger = log.New(os.Stdout, "normalizer:", log.LstdFlags)
}

func debugLog(msg string, args ...interface{}) {
	// A private, trivial trace logger
	if Debug {
		_, file1, pos1, _ := runtime.Caller(1)
		normLogger.Printf("%s:%d: %s", path.Base(file1), pos1, fmt.Sprintf(msg, args...))
	}
}
