package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
)

func TaskRunUsageTemplate() string {
	helpText := `
Usage: openstate task run <task> [options]

	Run an instance of the given task definition. This command initializes
	a copy of the task's underlying state machine and executable callbacks.
	To view a list of running task instances, use 'tasks ps'.

General Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`
	return strings.TrimSpace(helpText)
}

type TaskRunOptions struct {
	TaskName string
}

func NewTaskRunOptions() *TaskRunOptions {
	return &TaskRunOptions{}
}

func (o *TaskRunOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("'task run' takes exactly one argument.")
	}

	o.TaskName = args[0]
	return nil
}

func (o *TaskRunOptions) Run() {
	client, err := api.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Tasks().Run(o.TaskName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *res)
}

func NewCmdTaskRun() *cobra.Command {
	o := NewTaskRunOptions()

	cmd := &cobra.Command{
		Use: "run",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd, args); err != nil {
				fmt.Println(err)
				return
			}
			o.Run()
		},
	}

	cmd.SetUsageTemplate(TaskRunUsageTemplate())

	return cmd
}
