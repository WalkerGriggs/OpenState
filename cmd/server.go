package cmd

import (
	"net"
	"os"
	"strings"

	log "github.com/hashicorp/go-hclog"
	homedir "github.com/mitchellh/go-homedir"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/walkergriggs/openstate/openstate"
)

type ServerOptions struct {
	LogLevel        string
	SerfAdvertise   string
	RaftAdvertise   string
	HTTPAdvertise   string
	Peers           string
	NodeName        string
	ConfigPath      string
	BootstrapExpect int
}

func ServerUsageTemplate() string {
	helpText := `
Usage: openstate server [options]

	Run an OpenState server until and interupt is received or the server
	exists gracefully.

	The server is primarily configured with the config file, but the
	following flags may be used to overwrite config values.

General Options:

	--log-level=<level>
		The verbosity of OpenState's logger. Options include DEBUG, INFO,
		WARN, and ERROR. Defaults to INFO.

	--config=<path>
		Path to the config file. For the time being, this must be an
		absolute path. Defaults to $HOME/.openstate/config.yaml.

Server Options:

	--raft-address=<address>
		The host:port pair to serve the Raft endpoints on. Defaults to
		the loopback 127.0.0.1:7050.

	--serf-address=<address>
		The host:port pair to serve the Serf endpoints on. Defaults to
		the loopback 127.0.0.1:4648.

	--http-address=<address>
		The host:port pair to serve the HTTP endpoints on. Defaults to
		the loopback 127.0.0.1:8080.

	--join=<addresses>
		A comma separated list of Serf addresses to join. Only one valid
		address is necessary; the join event will be distributed to all
		currently-active peers.

	--bootstrap-expect=<N>
		The expected number of servers in the clusters. If 1, the server
		will bootstrap without any peers and elect itself the leader.
		Defaults to 1.

	--node-name=<name>
		Name given to the Serf node. This name is used to globally identify
		the node over the cluster's gossip protocol and must be unique.
		Defaults to the hostname.`

	return strings.TrimSpace(helpText)
}

func NewServerOptions() *ServerOptions {
	return &ServerOptions{
		LogLevel:        "INFO",
		BootstrapExpect: 1,
	}
}

func (o *ServerOptions) Complete(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *ServerOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *ServerOptions) extractToConfig(config *openstate.Config) error {
	config.Logger = log.NewInterceptLogger(&log.LoggerOptions{
		Name:   "OpenState",
		Level:  log.LevelFromString(o.LogLevel),
		Output: os.Stdout,
	})

	var err error

	// Extract the raft-address if provided.
	if o.RaftAdvertise != "" {
		config.RaftAdvertise, err = net.ResolveTCPAddr("tcp", o.RaftAdvertise)
		if err != nil {
			config.Logger.Error("Failed to resolve Raft address", "error", err.Error())
			return err
		}
	}

	// Extract the serf-address if provided.
	if o.SerfAdvertise != "" {
		config.SerfAdvertise, err = net.ResolveTCPAddr("tcp", o.SerfAdvertise)
		if err != nil {
			config.Logger.Error("Failed to resolve Serf address", "error", err.Error())
			return err
		}
	}

	// Extract the http-address if provided.
	if o.HTTPAdvertise != "" {
		config.HTTPAdvertise, err = net.ResolveTCPAddr("tcp", o.HTTPAdvertise)
		if err != nil {
			config.Logger.Error("Failed to resolve HTTP address", "error", err.Error())
			return err
		}
	}

	// Extract the node-name if provided.
	if o.NodeName != "" {
		config.NodeName = o.NodeName
	}

	// Extract bootstrap-expect and log-level
	config.BootstrapExpect = o.BootstrapExpect

	// Separate and set peer list.
	sep := strings.Split(o.Peers, ",")
	if len(sep) > 0 && len(sep[0]) > 0 {
		config.Peers = sep
	}

	return nil
}

func (o *ServerOptions) Run() {
	// Create default config
	config := openstate.DefaultConfig()

	// Read the config file and override server config defaults.
	if err := readConfig(o.ConfigPath); err != nil {
		return
	}

	extractConfig(config)

	// Override defaults and config file with flag values
	if err := o.extractToConfig(config); err != nil {
		return
	}

	// Configure a new server.
	server, err := openstate.NewServer(config)
	if err != nil {
		panic(err)
	}

	// Expose the server via HTTP endpoints.
	openstate.NewHTTPServer(server, config)

	// Off to the races!
	server.Run()
}

// NewCmdServer initializes ServerOptions, creates the new Cobra command, and
// adds the flags
func NewCmdServer() *cobra.Command {
	o := NewServerOptions()

	cmd := &cobra.Command{
		Use: "server",
		Run: func(cmd *cobra.Command, args []string) {
			o.Complete(cmd, args)
			o.Validate(cmd, args)
			o.Run()
		},
	}

	cmd.SetUsageTemplate(ServerUsageTemplate())

	cmd.Flags().StringVarP(&o.RaftAdvertise, "raft-address", "", o.RaftAdvertise, "")
	cmd.Flags().StringVarP(&o.SerfAdvertise, "serf-address", "", o.SerfAdvertise, "")
	cmd.Flags().StringVarP(&o.HTTPAdvertise, "http-address", "", o.HTTPAdvertise, "")
	cmd.Flags().StringVarP(&o.Peers, "join", "", o.Peers, "")
	cmd.Flags().StringVarP(&o.LogLevel, "log-level", "", o.LogLevel, "")
	cmd.Flags().StringVarP(&o.NodeName, "node-name", "", o.NodeName, "")
	cmd.Flags().StringVarP(&o.ConfigPath, "config", "", o.ConfigPath, "")
	cmd.Flags().IntVarP(&o.BootstrapExpect, "bootstrap-expect", "", o.BootstrapExpect, "")

	return cmd
}

// readConfig reads in the config file from $HOME/.openstate or the proided
// path.
func readConfig(path string) error {
	home, err := homedir.Dir()
	if err != nil {
		return err
	}

	if path == "" {
		viper.AddConfigPath(home + "/.openstate/")
		viper.SetConfigName("config")

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
	} else {
		// TODO don't hardcode the config type
		viper.SetConfigType("yaml")
		file, err := os.Open(path)
		if err != nil {
			return err
		}

		if err := viper.ReadConfig(file); err != nil {
			return err
		}
	}

	return nil
}

// extractConfig extracts and parses config values from Viper.
func extractConfig(config *openstate.Config) error {
	var err error

	raft_addr := viper.GetString("raft-address")
	if raft_addr != "" {
		config.RaftAdvertise, err = net.ResolveTCPAddr("tcp", raft_addr)
		if err != nil {
			return err
		}
	}

	serf_addr := viper.GetString("serf-address")
	if raft_addr != "" {
		config.SerfAdvertise, err = net.ResolveTCPAddr("tcp", serf_addr)
		if err != nil {
			return err
		}
	}

	http_addr := viper.GetString("http-address")
	if raft_addr != "" {
		config.HTTPAdvertise, err = net.ResolveTCPAddr("tcp", http_addr)
		if err != nil {
			return err
		}
	}

	node_name := viper.GetString("node-name")
	if node_name != "" {
		config.NodeName = node_name
	}

	bootstrap_expect := viper.GetInt("bootstrap-expect")
	if bootstrap_expect != 0 {
		config.BootstrapExpect = bootstrap_expect
	}

	config.Peers = viper.GetStringSlice("join")

	return nil
}
