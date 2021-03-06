package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/api"
)

type NamesOptions struct {
}

func NewNamesOptions() *NamesOptions {
	return &NamesOptions{}
}

func (o *NamesOptions) Complete(cmd *cobra.Command) error {
	return nil
}

func (o *NamesOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *NamesOptions) Run() {
	client, err := api.NewClient()
	if err != nil {
		panic(err)
	}

	names, err := client.Names().List()
	if err != nil {
		panic(err)
	}

	fmt.Println(names)
}

func NewCmdNames() *cobra.Command {
	o := NewNamesOptions()

	cmd := &cobra.Command{
		Use:   "names",
		Short: "List names persisted to OpenState's Raft FSM.",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(cmd)
			o.Validate(cmd, args)
			o.Run()
		},
	}

	return cmd
}
