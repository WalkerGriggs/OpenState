package cmd

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	log "github.com/hashicorp/go-hclog"
	"github.com/mitchellh/cli"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"

	"github.com/walkergriggs/openstate/openstate"
)

func ServerUsageTemplate() string {
	helpText := `
Usage: openstate server [options]

	Run an OpenState server until and interupt is received or the server
	exists gracefully.

	The server is primarily configured with the config file, but the
	following flags may be used to overwrite config values.

General Options:

	--dev
		Run the server in development mode. This disables the log, stable,
		and snapshot stores, and uses a simple in-memory store instead.

	--log-level=<level>
		The verbosity of OpenState's logger. Options include DEBUG, INFO,
		WARN, and ERROR. Defaults to INFO.

	--config=<path>
		Path to the config file. For the time being, this must be an
		absolute path. Defaults to $HOME/.openstate/config.yaml.

	--data-dir=<path>
		Path to directory where OpenState stores state related objects;
		primarily snapshots, logs, and the stable store.

	--no-banner
		Do not print the OpenState ascii banner

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

// ServerOptions wraps the Config and any additional flags needed to run a new
// server.
type ServerOptions struct {
	Meta
	config     *Config
	configPath string
}

func NewServerOptions() *ServerOptions {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	pui := &cli.PrefixedUi{
		OutputPrefix: "... ",
		InfoPrefix:   "    ",
		ErrorPrefix:  "... ",
		Ui:           ui,
	}

	return &ServerOptions{
		Meta: Meta{UI: pui},
		config: &Config{
			Addrs:  &AdvertiseAddrs{},
			Server: &ServerConfig{},
		},
	}
}

func (o *ServerOptions) Complete(cmd *cobra.Command, args []string) error {
	if o.configPath == "" {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}

		o.configPath = path.Join(home, CONFIG_DIR, CONFIG_FILENAME)
	}

	return nil
}

// Run reads in the config file, overwrites any values with flags, and starts
// the server.
func (o *ServerOptions) Run() {
	if !o.config.NoBanner {
		o.Meta.UI.Info(Banner(0))
	}

	// Read the config file and unmarshal the results
	config, err := unmarshalConfig(o.configPath, o.Meta.UI)
	if err != nil {
		o.UI.Error(err.Error())
		return
	}

	// Overwrite config with command flags
	config = config.merge(o.config)

	// Convert cmd.Config to openstate.Config
	serverConfig, err := config.ctoc()
	if err != nil {
		o.UI.Error(err.Error())
		return
	}

	// Set the logger
	serverConfig.Logger = log.NewInterceptLogger(&log.LoggerOptions{
		Name:   "OpenState",
		Level:  log.LevelFromString(config.LogLevel),
		Output: serverConfig.LogOutput,
	})

	info := make(map[string]string)
	info["Log Level"] = config.LogLevel
	info["Dev Mode"] = strconv.FormatBool(serverConfig.DevMode)
	info["Advertise Addrs"] = fmt.Sprintf("HTTP: %s Serf: %s Raft: %s",
		serverConfig.HTTPAdvertise.String(),
		serverConfig.SerfAdvertise.String(),
		serverConfig.RaftAdvertise.String(),
	)

	o.Meta.UI.Output("Server configuration:\n")

	for k, v := range info {
		o.Meta.UI.Info(fmt.Sprintf(
			"%s%s: %s",
			strings.Repeat(" ", 20-len(k)),
			strings.Title(k), v))
	}

	o.Meta.UI.Output("")
	o.Meta.UI.Output("Starting OpenState server:\n")

	// Create the new server
	server, err := openstate.NewServer(serverConfig)
	if err != nil {
		o.Meta.UI.Error(err.Error())
		return
	}

	// Wrap the server and expose it over HTTP endpoints
	openstate.NewHTTPServer(server, serverConfig)

	// Off to the races!
	server.Run()
}

// NewCmdServer initializes ServerOptions, creates the new Cobra command, and
// adds the flags
func NewCmdServer() *cobra.Command {
	o := NewServerOptions()
	config := o.config

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Runs an OpenState server",
		Run: func(cmd *cobra.Command, args []string) {
			if err := o.Complete(cmd, args); err != nil {
				return
			}

			o.Run()
		},
	}

	cmd.SetUsageTemplate(ServerUsageTemplate())

	// Address Flags
	cmd.Flags().StringVar(&config.Addrs.Raft, "raft-address", config.Addrs.Raft, "")
	cmd.Flags().StringVar(&config.Addrs.Serf, "serf-address", config.Addrs.Serf, "")
	cmd.Flags().StringVar(&config.Addrs.HTTP, "http-address", config.Addrs.HTTP, "")

	// Server Flags
	cmd.Flags().StringSliceVar(&config.Server.Join, "join", config.Server.Join, "")
	cmd.Flags().StringVar(&config.Server.NodeName, "node-name", config.Server.NodeName, "")
	cmd.Flags().IntVar(&config.Server.BootstrapExpect, "bootstrap-expect", config.Server.BootstrapExpect, "")

	// General Flags
	cmd.Flags().BoolVar(&config.NoBanner, "no-banner", config.NoBanner, "")
	cmd.Flags().BoolVar(&config.DevMode, "dev", config.DevMode, "")
	cmd.Flags().StringVar(&config.LogLevel, "log-level", config.LogLevel, "")
	cmd.Flags().StringVar(&config.DataDirectory, "data-dir", config.DataDirectory, "")
	cmd.Flags().StringVar(&o.configPath, "config", o.configPath, "")

	return cmd
}
