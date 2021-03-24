package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func TaskListUsageTemplate() string {
	helpText := `
Usage: openstate task list [options]

	List all currently defined tasks.

` + SharedUsageTemplate()

	return strings.TrimSpace(helpText)
}

type TaskListOptions struct {
	Meta
}

func NewTaskListOptions() *TaskListOptions {
	return &TaskListOptions{}
}

func (o *TaskListOptions) Run() {
	client, err := o.Meta.Client()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Tasks().List()
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *res)
}

func (o *TaskListOptions) Name() string {
	return "task list"
}

func NewCmdTaskList() *cobra.Command {
	o := NewTaskListOptions()

	cmd := &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			o.Run()
		},
	}

	sharedFlags := o.Meta.FlagSet(o.Name())

	cmd.Flags().AddFlagSet(sharedFlags)
	cmd.SetUsageTemplate(TaskListUsageTemplate())

	return cmd
}
