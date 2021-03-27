package cmd

import (
	"strings"

	"github.com/spf13/pflag"

	"github.com/walkergriggs/openstate/api"
)

type Meta struct {
	UI      UI
	Address string
}

// clientConfig returns a default api.Config with optional shared flag values
// merged in.
func (m *Meta) clientConfig() *api.Config {
	config := api.DefaultConfig()

	if m.Address != "" {
		config.Address = m.Address
	}

	return config
}

// Client is a wrapper around api.NewClient that merges in optional client
// configs from the shared CLI flags.
func (m *Meta) Client() (*api.Client, error) {
	return api.NewClient(m.clientConfig())
}

// SharedUsageTemplate retuns a formatted string enumerating each shared CLI
// flag. It should not be used as an argument to SetUsageTemplate, but instead
// factored into a command's own usage template.
func SharedUsageTemplate() string {
	helpText := `
Shared Options:

	--address=<address>
		The host:port pair of an OpenState server HTTP endpoint. This
		endpoint can be any server in the cluster; the request will be
		forwarded to the leader.
`

	return strings.TrimSpace(helpText)
}

// FlagSet returns posix-style FlagSet for all shared CLI flags. This should be
// merged into a command's own FlagSet.
func (m *Meta) FlagSet(n string) *pflag.FlagSet {
	f := pflag.NewFlagSet(n, pflag.ContinueOnError)

	f.StringVar(&m.Address, "address", m.Address, "")

	return f
}
