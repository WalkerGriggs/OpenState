package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdInstance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "Groups instance-related subcommands",
	}

	cmd.AddCommand(NewCmdInstanceEvent())

	return cmd
}
