package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/spf13/cobra"
)

func TaskRunUsageTemplate() string {
	helpText := `
Usage: openstate task run <task> [options]

	Run an instance of the given task definition. This command initializes
	a copy of the task's underlying state machine and executable callbacks.
	To view a list of running task instances, use 'tasks ps'.

` + SharedUsageTemplate()

	return strings.TrimSpace(helpText)
}

type TaskRunOptions struct {
	Meta
	TaskName string
}

func NewTaskRunOptions() *TaskRunOptions {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	return &TaskRunOptions{
		Meta: Meta{UI: ui},
	}
}

func (o *TaskRunOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("'task run' takes exactly one argument.")
	}

	o.TaskName = args[0]
	return nil
}

func (o *TaskRunOptions) Run() {
	client, err := o.Meta.Client()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Tasks().Run(o.TaskName)
	if err != nil {
		fmt.Println(err)
		return
	}

	o.Meta.UI.Output(res.Instance.Summarize().String())
}

func (o *TaskRunOptions) Name() string {
	return "task run"
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

	sharedFlags := o.Meta.FlagSet(o.Name())

	cmd.Flags().AddFlagSet(sharedFlags)
	cmd.SetUsageTemplate(TaskRunUsageTemplate())

	return cmd
}
