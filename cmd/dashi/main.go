package main

import (
	"os"

	"github.com/cmessinides/dashi/internal/cli"
)

var DevMode = "off"

func main() {
	os.Exit(cli.Run(cli.Config{
		Dev: DevMode == "on",
	}))
}
