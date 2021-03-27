package cmd

import (
	"fmt"
	"io"
)

// UI describes the behavior of a CLI UI
type UI interface {
	Output(string)
}

// SimpleUI implements UI with base functionality.
type SimpleUI struct {
	Writer io.Writer
}

// Output prints the given string over the SimpleUI's writer. Aside from
// appending a newline character, it does not format the string in any way.
func (ui *SimpleUI) Output(s string) {
	fmt.Fprint(ui.Writer, s)
	fmt.Fprint(ui.Writer, "\n")
}

// PrefixedUI is a wrapper for the UI interface. It prefixes is non-zero string
// with the configured prefix.
type PrefixedUI struct {
	UI
	Prefix string
}

// Ouput prefixes the given string (if non-zero) and passes it to the underlying
// UI's Output method.
func (ui *PrefixedUI) Output(s string) {
	if s != "" {
		s = fmt.Sprintf("%s%s", ui.Prefix, s)
	}

	ui.UI.Output(s)
}
