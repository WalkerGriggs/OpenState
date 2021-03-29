package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/spf13/cobra"
)

func TaskPsUsageTemplate() string {
	helpText := `
Usage: openstate task ps <task> [options]

	List all running instances of the defined task.

` + SharedUsageTemplate()

	return strings.TrimSpace(helpText)
}

type TaskPsOptions struct {
	Meta
	TaskName string
}

func NewTaskPsOptions() *TaskPsOptions {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	return &TaskPsOptions{
		Meta: Meta{UI: ui},
	}
}

func (o *TaskPsOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("'task ps' takes exactly one argument.")
	}

	o.TaskName = args[0]
	return nil
}

func (o *TaskPsOptions) Run() {
	client, err := o.Meta.Client()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Tasks().Ps(o.TaskName)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *res)
}

func (o *TaskPsOptions) Name() string {
	return "task list"
}

func NewCmdTaskPs() *cobra.Command {
	o := NewTaskPsOptions()

	cmd := &cobra.Command{
		Use: "ps",
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
	cmd.SetUsageTemplate(TaskPsUsageTemplate())

	return cmd
}
