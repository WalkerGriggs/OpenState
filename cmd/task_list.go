package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
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
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	return &TaskListOptions{
		Meta: Meta{UI: ui},
	}
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

	table := FormatTable(formatDefinitions(res.Definitions))

	o.UI.Output(table)
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

func formatDefinitions(definitions []*api.Definition) (data [][]string) {
	data = append(data, []string{"Name"})

	for _, def := range definitions {
		data = append(data, []string{def.Metadata.Name})
	}

	return
}
