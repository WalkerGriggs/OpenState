package cmd

import (
	"github.com/spf13/cobra"
)

func NewCmdTask() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "task groups task-related subcommands",
	}

	cmd.AddCommand(NewCmdTaskDefine())
	cmd.AddCommand(NewCmdTaskList())

	return cmd
}
