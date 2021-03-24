package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var SharedInstanceOptions *InstanceOptions = &InstanceOptions{}

type InstanceOptions struct {
	Address string
}

func SharedInstanceUsageTemplate() string {
	helpText := `
Shared Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`
	return strings.TrimSpace(helpText)
}

func NewCmdInstance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance",
		Short: "instance groups instance-related subcommands",
	}

	cmd.AddCommand(NewCmdInstanceEvent())

	cmd.PersistentFlags().StringVar(&SharedInstanceOptions.Address, "address", SharedInstanceOptions.Address, "")

	return cmd
}
