package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
)

func TaskPsUsageTemplate() string {
	helpText := `
Usage: openstate task ps <task> [options]

	List all running instances of the defined task.

` + SharedTaskUsageTemplate()

	return strings.TrimSpace(helpText)
}

type TaskPsOptions struct {
	TaskName string
}

func NewTaskPsOptions() *TaskPsOptions {
	return &TaskPsOptions{}
}

func (o *TaskPsOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("'task ps' takes exactly one argument.")
	}

	o.TaskName = args[0]
	return nil
}

func (o *TaskPsOptions) Run() {
	client, err := api.NewClient()
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

	cmd.SetUsageTemplate(TaskPsUsageTemplate())

	return cmd
}
