package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/walkergriggs/openstate/api"
)

func TaskDefineUsageTemplate() string {
	helpText := `
Usage: openstate task define <path> [options]

	Define a new or update an existing Task using the definition file
	at <path>. For the time being, this path must be absolute.

` + SharedTaskUsageTemplate()

	return strings.TrimSpace(helpText)
}

type TaskDefineOptions struct {
	path string
}

func NewTaskDefineOptions() *TaskDefineOptions {
	return &TaskDefineOptions{}
}

func (o *TaskDefineOptions) Complete(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("task define takes exactly 1 argument")
	}

	o.path = args[0]
	return nil
}

func (o *TaskDefineOptions) Run() {
	f, err := os.Open(o.path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}

	def := &api.Definition{
		FSM: &api.FSM{
			Events: make([]*api.Event, 0),
		},
	}

	yaml.Unmarshal(b, def)

	client, err := api.NewClient()
	if err != nil {
		fmt.Println(err)
		return
	}

	res, err := client.Tasks().Define(def)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%+v\n", *res)
}

func NewCmdTaskDefine() *cobra.Command {
	o := NewTaskDefineOptions()

	cmd := &cobra.Command{
		Use: "define",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd, args); err != nil {
				fmt.Println(err.Error())
				return
			}
			o.Run()
		},
	}

	cmd.SetUsageTemplate(TaskDefineUsageTemplate())

	return cmd
}
