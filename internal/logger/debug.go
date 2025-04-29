package logger

import (
	"log"
)

// Enabled controls whether debug logs are printed.
var DebugEnabled bool

// Debug prints a message if debug mode is enabled.
func Debug(format string, v ...interface{}) {
	if DebugEnabled {
		log.Printf(format, v...)
	}
}
