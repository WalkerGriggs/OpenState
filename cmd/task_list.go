package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
)

func TaskListUsageTemplate() string {
	helpText := `
Usage: openstate task list [options]

	List all currently defined tasks.

General Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`

	return strings.TrimSpace(helpText)
}

type TaskListOptions struct {
}

func NewTaskListOptions() *TaskListOptions {
	return &TaskListOptions{}
}

func (o *TaskListOptions) Run() {
	client, err := api.NewClient()
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

func NewCmdTaskList() *cobra.Command {
	o := NewTaskListOptions()

	cmd := &cobra.Command{
		Use: "list",
		Run: func(cmd *cobra.Command, args []string) {
			o.Run()
		},
	}

	cmd.SetUsageTemplate(TaskListUsageTemplate())

	return cmd
}