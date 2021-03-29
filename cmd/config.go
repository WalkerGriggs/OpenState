package cmd

import (
	"fmt"
	"net"
	"path"
	"strings"

	"github.com/mitchellh/cli"
	"github.com/spf13/viper"

	"github.com/walkergriggs/openstate/openstate"
)

const (
	CONFIG_DIR = ".openstate"

	CONFIG_FILENAME = "config"
)

type (
	// Config is the umbrella struct with encompasses every configurable value in
	// OpenState's config file.
	Config struct {
		// NoBanner indicates if we should not print the ASCII OpenState banner
		NoBanner bool `mapstructure:"no_banner"`

		// DevMode indicates if the OpenState server should run in development mode.
		// This disables the log, stable, and snapshot stores, and uses a simple
		// in-memory store instead.
		DevMode bool `mapstructure:"dev_mode"`

		// DataDirectory is a path to directory where OpenState stores state related
		// objects; primarily snapshots, logs, and the stable store.
		DataDirectory string `mapstructure:data_directory`

		// LogLevel is the verbosity of OpenState's logger. Options include: DEBUG,
		// INFO, WARN, and ERROR. See hashicorp/hclog for more info.
		LogLevel string `mapstructure:"log_level"`

		// Addrs are the addresses for Raft, Serf, and HTTP endpoints.
		Addrs *AdvertiseAddrs `mapstructure:"advertise"`

		// Server contains server-specific configs.
		Server *ServerConfig `yaml:"server"`
	}

	// AdvertiseAddrs are the addresses for Raft, Serf, and HTTP endpoints.
	AdvertiseAddrs struct {
		HTTP string `mapstructure:"http"`
		Serf string `mapstructure:"serf"`
		Raft string `mapstructure:"raft"`
	}

	// ServerConfig contains server-specific configs.
	ServerConfig struct {
		// NodeNameis the advertised name of the server. It's most commonly used in
		// Serf's gossip protocol.
		NodeName string `mapstructure:"node_name"`

		// Join is the initial list of peer serf addresses. This option is a hack
		// to bypass the need for service discovery (TODO). This list only needs to
		// contain ONE valid peer; the gossip layer will propogate the new peer
		// across all nodes.
		Join []string `mapstructure:"join"`

		// BootstrapExpect is the number of peers the new server should expect. If
		// 1, the new server will create a single node cluster and elect itself
		// leader.
		BootstrapExpect int `mapstructure:"bootstrap_expect"`
	}
)

// merge sets values from the argument Config (b) to fields in the receiving
// Config (a) and returns the merged result.
func (a *Config) merge(b *Config) *Config {
	res := *a

	if b.LogLevel != "" {
		res.LogLevel = b.LogLevel
	}

	if b.DevMode != false {
		res.DevMode = b.DevMode
	}

	if b.DataDirectory != "" {
		res.DataDirectory = b.DataDirectory
	}

	// Set the server configs, or merge if necessary.
	if res.Server == nil && b.Server != nil {
		server := *b.Server
		res.Server = &server
	} else if b.Server != nil {
		res.Server = res.Server.merge(b.Server)
	}

	// Set the addresses, or merge if necessary.
	if res.Addrs == nil && b.Addrs != nil {
		addrs := *b.Addrs
		res.Addrs = &addrs
	} else if b.Addrs != nil {
		res.Addrs = res.Addrs.merge(b.Addrs)
	}

	return &res
}

// merge sets values from the argument ServerConfig (b) to fields in the
// receiving ServerConfig (a) and returns the merged result.
func (a *ServerConfig) merge(b *ServerConfig) *ServerConfig {
	res := *a

	if b.NodeName != "" {
		res.NodeName = b.NodeName
	}

	if b.Join != nil {
		res.Join = b.Join
	}

	if b.BootstrapExpect > 0 {
		res.BootstrapExpect = b.BootstrapExpect
	}

	return &res
}

// merge sets values from the argument AdvertiseAddrs (b) to fields in the
// receiving AdvertiseAddrs (a) and returns the merged result.
func (a *AdvertiseAddrs) merge(b *AdvertiseAddrs) *AdvertiseAddrs {
	res := *a

	if b.Raft != "" {
		res.Raft = b.Raft
	}

	if b.Serf != "" {
		res.Serf = b.Serf
	}

	if b.HTTP != "" {
		res.HTTP = b.HTTP
	}

	return &res
}

// ctoc converts the cmd.Config to an openstate.Config
func (conf *Config) ctoc() (*openstate.Config, error) {
	serverConf := openstate.DefaultConfig()

	if conf.Server.NodeName != "" {
		serverConf.NodeName = conf.Server.NodeName
	}

	if conf.Server.BootstrapExpect > 0 {
		serverConf.BootstrapExpect = conf.Server.BootstrapExpect
	}

	if conf.Server.Join != nil {
		serverConf.Peers = conf.Server.Join
	}

	if conf.DevMode != false {
		serverConf.DevMode = conf.DevMode
	}

	if conf.DataDirectory != "" {
		serverConf.DataDirectory = conf.DataDirectory
	}

	var err error

	if conf.Addrs.Raft != "" {
		serverConf.RaftAdvertise, err = net.ResolveTCPAddr("tcp", conf.Addrs.Raft)
		if err != nil {
			return nil, err
		}
	}

	if conf.Addrs.Serf != "" {
		serverConf.SerfAdvertise, err = net.ResolveTCPAddr("tcp", conf.Addrs.Serf)
		if err != nil {
			return nil, err
		}
	}

	if conf.Addrs.HTTP != "" {
		serverConf.HTTPAdvertise, err = net.ResolveTCPAddr("tcp", conf.Addrs.HTTP)
		if err != nil {
			return nil, err
		}
	}

	return serverConf, nil
}

// unmarshalConfig reads in the config file from either the default directory
// or the given path and returns the extracted values in a Config struct.
func unmarshalConfig(p string, ui cli.Ui) (*Config, error) {
	viper.BindEnv("dev_mode")
	viper.BindEnv("log_level")
	viper.BindEnv("data_directory")
	viper.BindEnv("advertise.http")
	viper.BindEnv("advertise.raft")
	viper.BindEnv("advertise.serf")
	viper.BindEnv("server.node_name")
	viper.BindEnv("server.join")
	viper.BindEnv("server.bootstrap_expect")
	viper.AutomaticEnv()

	configPath, configFile := path.Split(p)
	configName := strings.TrimSuffix(configFile, path.Ext(p))

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetEnvPrefix("OS")

	viper.AddConfigPath(configPath)
	viper.SetConfigName(configName)

	ui.Output(fmt.Sprintf("Checking for config file '%s.*' in %s", configName, configPath))

	if err := viper.ReadInConfig(); err != nil {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			ui.Output("No config file found. Falling back to environment variables.")
		default:
			return nil, err
		}
	}

	config := &Config{
		Addrs:  &AdvertiseAddrs{},
		Server: &ServerConfig{},
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
