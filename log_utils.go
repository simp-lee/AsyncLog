package asynclog

import (
	"path/filepath"
	"runtime"
)

// getCallerInfo retrieves the filename and line number of the log caller.
func getCallerInfo() (string, int) {
	_, file, line, ok := runtime.Caller(3) // Adjust the stack frame to get the correct caller
	if !ok {
		return "unknown", 0
	}
	return filepath.Base(file), line
}
