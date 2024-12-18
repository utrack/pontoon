package httpinmeditate

import (
	"fmt"
)

var debugEnabled = false

func debugLog(f string, args ...interface{}) {
	if debugEnabled {
		fmt.Printf(f+"\n", args...)
	}
}
