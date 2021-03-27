package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func InstanceEventUsageTemplate() string {
	helpText := `
Usage: openstate instance event <instance> <event> [options]

	Perform an event against a specific task instance. This command returns
	the current state of the instance after the even is performed. If the
	instance cannot perform the event, it will return an error.

` + SharedUsageTemplate()

	return strings.TrimSpace(helpText)
}

type InstanceEventOptions struct {
	Meta
	instanceName string
	eventName    string
}

func NewInstanceEventOptions() *InstanceEventOptions {
	ui := &SimpleUI{os.Stdout}

	return &InstanceEventOptions{
		Meta: Meta{UI: ui},
	}
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
	client, err := o.Meta.Client()
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

func (o *InstanceEventOptions) Name() string {
	return "instance event"
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

	sharedFlags := o.Meta.FlagSet(o.Name())

	cmd.Flags().AddFlagSet(sharedFlags)
	cmd.SetUsageTemplate(InstanceEventUsageTemplate())

	return cmd
}
