package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var SharedTaskOptions *TaskOptions = &TaskOptions{}

type TaskOptions struct {
	Address string
}

func SharedTaskUsageTemplate() string {
	helpText := `
Shared Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`

	return strings.TrimSpace(helpText)
}

func NewCmdTask() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "task groups task-related subcommands",
	}

	cmd.AddCommand(NewCmdTaskDefine())
	cmd.AddCommand(NewCmdTaskList())
	cmd.AddCommand(NewCmdTaskRun())
	cmd.AddCommand(NewCmdTaskPs())
	// cmd.AddCommand(NewCmdTaskInspect())

	cmd.PersistentFlags().StringVar(&SharedTaskOptions.Address, "address", SharedTaskOptions.Address, "")

	return cmd
}
