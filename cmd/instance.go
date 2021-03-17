package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdInstance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "instance groups instance-related subcommands",
	}

	cmd.AddCommand(NewCmdInstanceEvent())

	return cmd
}
