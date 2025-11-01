package debug

import (
	"fmt"
	"os"
)

// DebugEnabled controls whether debug messages are printed to stderr.
// It can be set directly or by setting the TERMPLEX_DEBUG=1 environment variable.
var DebugEnabled = false

func init() {
	if os.Getenv("TERMPLEX_DEBUG") == "1" {
		DebugEnabled = true
	}
}

// Log prints a formatted debug message to stderr if debugging is enabled.
func Log(format string, a ...interface{}) {
	if DebugEnabled {
		fmt.Fprintf(os.Stderr, "[TERMPLEX DEBUG] "+format+"\n", a...)
	}
}
