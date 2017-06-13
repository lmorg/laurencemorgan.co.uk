// functions
package main

import (
	"strings"
)

func windowsfyPath(s string) string {
	win:= strings.Replace(s, "/", `\`, -1)
	debugLog("[functions_windows] [windowsfyPath]", win)
	return win
}
