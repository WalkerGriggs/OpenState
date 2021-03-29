package cmd

import (
	"strings"
)

func Banner(padding int) string {
	ascii :=
		`____                  _____ __        __
   / __ \____  ___  ____ / ___// /_____ _/ /____
  / / / / __ \/ _ \/ __ \\__ \/ __/ __ '/ __/ _ \
 / /_/ / /_/ /  __/ / / /__/ / /_/ /_/ / /_/  __/
 \____/ .___/\___/_/ /_/____/\__/\__,_/\__/\___/
     /_/
`

	split := strings.Split(ascii, "\n")

	pad := strings.Repeat(" ", padding)

	return strings.Join(split[:], "\n"+pad)
}
