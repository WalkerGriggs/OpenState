package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdOpenState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "openstate",
		Short: "Language agnostic task runner",
	}

	cmd.AddCommand(NewCmdServer())
	cmd.AddCommand(NewCmdNames())

	return cmd
}
