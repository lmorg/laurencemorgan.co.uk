package main

import (
	"github.com/kardianos/osext"
)

func defaultConfDir() string {
	path, err := osext.ExecutableFolder()
	failOnErr(err, "defaultConfDir (windows)")
	return path + `\..\conf\`
}
