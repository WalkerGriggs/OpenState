package cmd

import (
	"net"
	"os"
	"strings"

	log "github.com/hashicorp/go-hclog"

	"github.com/spf13/cobra"
	"github.com/walkergriggs/openstate/openstate"
)

type ServerOptions struct {
	LogLevel      string
	SerfAdvertise string
	RaftAdvertise string
	Peers         string
	SerfName      string
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{}
}

func (o *ServerOptions) Complete(cmd *cobra.Command) error {
	if o.LogLevel == "" {
		o.LogLevel = "INFO"
	}

	return nil
}

func (o *ServerOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *ServerOptions) Run() {
	logger := log.NewInterceptLogger(&log.LoggerOptions{
		Name:   "OpenState",
		Level:  log.LevelFromString(o.LogLevel),
		Output: os.Stdout,
	})

	config := openstate.DefaultConfig()
	config.Logger = logger

	var err error

	if o.RaftAdvertise != "" {
		config.RaftAdvertise, err = net.ResolveTCPAddr("tcp", o.RaftAdvertise)
		if err != nil {
			panic(err)
		}
	}

	if o.SerfAdvertise != "" {
		config.SerfAdvertise, err = net.ResolveTCPAddr("tcp", o.SerfAdvertise)
		if err != nil {
			panic(err)
		}
	}

	if o.SerfName != "" {
		config.NodeName = o.SerfName
	}

	sep := strings.Split(o.Peers, ",")
	if len(sep) > 0 && len(sep[0]) > 0 {
		config.Peers = sep
		config.BootstrapExpect = len(sep)
	}

	server, err := openstate.NewServer(config)
	if err != nil {
		panic(err)
	}

	server.Run()
}

func NewCmdServer() *cobra.Command {
	o := NewServerOptions()

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Subcommand for interacting with OpenState servers.",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(cmd)
			o.Validate(cmd, args)
			o.Run()
		},
	}

	cmd.Flags().StringVarP(&o.RaftAdvertise, "raft_addr", "r", o.RaftAdvertise, "Advertise address for Raft")
	cmd.Flags().StringVarP(&o.SerfAdvertise, "serf_addr", "s", o.SerfAdvertise, "Advertise address for Serf")
	cmd.Flags().StringVarP(&o.Peers, "peers", "p", o.Peers, "Comma seperated list of peers.")
	cmd.Flags().StringVarP(&o.LogLevel, "level", "l", o.LogLevel, "Log level [DEBUG, INFO, WARN, ERROR].")
	cmd.Flags().StringVarP(&o.SerfName, "serf_name", "n", o.SerfName, "Node name for Serf cluster. Defaults to hostname.")

	return cmd
}
