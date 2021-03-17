package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
)

func InstanceEventUsageTemplate() string {
	helpText := `
Usage: openstate instance event <instance> <event> [options]

	Perform an event against a specific task instance. This command returns
	the current state of the instance after the even is performed. If the
	instance cannot perform the event, it will return an error.

General Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`
	return strings.TrimSpace(helpText)
}

type InstanceEventOptions struct {
	instanceName string
	eventName    string
}

func NewInstanceEventOptions() *InstanceEventOptions {
	return &InstanceEventOptions{}
}

func (o *InstanceEventOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("instance event takes exactly 2 arguments")
	}

	o.instanceName = args[0]
	o.eventName = args[1]
	return nil
}

func (o *InstanceEventOptions) Run() {
	client, err := api.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Instances().Event(o.instanceName, o.eventName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *res)
}

func NewCmdInstanceEvent() *cobra.Command {
	o := NewInstanceEventOptions()

	cmd := &cobra.Command{
		Use: "event",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd, args); err != nil {
				fmt.Println(err)
				return
			}
			o.Run()
		},
	}

	cmd.SetUsageTemplate(InstanceEventUsageTemplate())

	return cmd
}
