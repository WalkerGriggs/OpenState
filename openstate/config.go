package openstate

import (
	"io"
	"net"
	"os"
	"time"

	log "github.com/hashicorp/go-hclog"

	"github.com/google/uuid"
	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

const (
	DefaultRaftPort = 7050
	DefaultSerfPort = 4648
	DefaultHTTPPort = 8080
)

// Config is the comprehensive list of Server options.
type Config struct {
	// BootstrapExpect is used to determine how many peers to expect.
	//
	// The BootstrapExpect can be of any of the following values:
	//  1: Server will form a single node cluster and become a leader immediately
	//  N, larger than 1: Server will wait until it's connected to N servers
	//      before attempting leadership and forming the cluster.  No Raft Log operation
	//      will succeed until then.
	//
	// Defaults to 1
	BootstrapExpect int

	// DevMode indicates if the server is run in development mode. Dev mode limits
	// persistence and state to in-memory.
	//
	// Defaults to 'false'
	DevMode bool

	// HTTPAdvertise is the advertised address of the HTTP endpoints.
	//
	// Defaults to "127.0.0.1:8080"
	HTTPAdvertise *net.TCPAddr

	// Logger is the logger used by the OpenState server, raft, and serf.
	Logger log.InterceptLogger

	// LogOutput is the location to write logs to.
	//
	// Defaults to stdout.
	LogOutput io.Writer

	// NodeID is the UUID of the server
	//
	// Defaults to a random UUID
	NodeID string

	// NodeName is the advertised name of the server.
	//
	// Defaults to the node's hostname.
	NodeName string

	// HACK / TODO / REMOVE ME
	//
	// Peers are the initial list of peer serf addresses. This option is a hack
	// to bypass the need for service discovery (TODO). This list only needs to
	// contain ONE valid peer; the gossip layer will propogate the peer across all
	// nodes.
	//
	// Defaults to an empty list
	Peers []string

	// RaftAdvertise is the advertised address of the Raft node. This should
	// differ from the SerfAdvertise.
	//
	// Defaults to "0.0.0.0:5479"
	RaftAdvertise *net.TCPAddr

	// RaftConfig is the configuration used for Raft.
	RaftConfig *raft.Config

	// RaftTimeout is applied to any network traffic for raft.
	//
	// Defaults to 10s.
	RaftTimeout time.Duration

	// SerfAdvertise is the advertised address of the Serf node. This should
	// differ from the RaftAdvertise.
	//
	// Defaults to "0.0.0.0:7479"
	SerfAdvertise *net.TCPAddr

	// SerfConfig is the configuration used for Serf.
	SerfConfig *serf.Config
}

func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	c := &Config{
		BootstrapExpect: 1,
		DevMode:         false,
		HTTPAdvertise:   DefaultHTTPAddr(),
		LogOutput:       os.Stdout,
		NodeID:          generateUUID(),
		NodeName:        hostname,
		Peers:           make([]string, 0),
		RaftAdvertise:   DefaultRaftAddr(),
		RaftConfig:      raft.DefaultConfig(),
		RaftTimeout:     10 * time.Second,
		SerfAdvertise:   DefaultSerfAddr(),
		SerfConfig:      serf.DefaultConfig(),
	}

	c.SerfConfig.MemberlistConfig = memberlist.DefaultWANConfig()
	c.SerfConfig.MemberlistConfig.BindPort = DefaultSerfPort
	c.SerfConfig.MemberlistConfig.AdvertisePort = DefaultSerfPort

	c.RaftConfig.ShutdownOnRemove = false

	return c
}

func DefaultRaftAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: DefaultRaftPort,
	}
}

func DefaultSerfAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: DefaultSerfPort,
	}
}

func DefaultHTTPAddr() *net.TCPAddr {
	return &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: DefaultHTTPPort,
	}
}

func generateUUID() string {
	return uuid.Must(uuid.NewRandom()).String()
}
