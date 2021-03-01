package main

import (
	"os"

	"github.com/walkergriggs/openstate/cmd"
)

func main() {
	command := cmd.NewCmdOpenState()

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
