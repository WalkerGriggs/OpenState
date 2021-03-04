package openstate

import (
	"fmt"
	"time"

	log "github.com/hashicorp/go-hclog"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

// Server is the OpenState server which manages tasks etc.
type Server struct {
	// Server configuration with reasonable defaults.
	config *Config

	// eventCh is used to receive Serf events.
	eventCh chan serf.Event

	// Finite State Machine used to maintain state across Raft nodes
	fsm *FSM

	// logger is an hclog instance to better interact with Hashi's Raft config
	logger log.InterceptLogger

	// peers tracks known OpenState servers
	peers []string

	// raft is used for strong consistency and replicated state
	raft      *raft.Raft
	raftInmem *raft.InmemStore

	// reconcileCh is used to pass membership events between the Serf gossip layer
	// and the leader management loop.
	reconcileCh chan serf.Member

	// serf is a gossip agent for membership tracking and health checks
	serf *serf.Serf
}

func NewServer(c *Config) (*Server, error) {
	s := &Server{
		config:      c,
		logger:      c.Logger,
		eventCh:     make(chan serf.Event, 256),
		reconcileCh: make(chan serf.Member, 32),
	}

	var err error

	s.raft, err = s.setupRaft()
	if err != nil {
		return nil, fmt.Errorf("Failed to start Raft: %v", err)
	}

	s.serf, err = s.setupSerf()
	if err != nil {
		return nil, fmt.Errorf("Failed to start Serf: %v", err)
	}

	// Handle Serf events
	go s.serfEventHandler()

	// Monitor as assume leadership responsibilities
	go s.monitorLeadership()

	// Bootstrap the new server
	go s.bootstrapHandler(c.Peers)

	return s, nil
}

// setupRaft sets up and initializes the Raft node.
func (s *Server) setupRaft() (*raft.Raft, error) {
	config := s.config.RaftConfig

	// Update configs
	config.Logger = s.logger
	config.LocalID = raft.ServerID(s.config.NodeID)

	// Initialize the server's FSM
	fsmConfig := &FSMConfig{
		Logger: s.logger,
	}

	var err error
	s.fsm, err = NewFSM(fsmConfig)
	if err != nil {
		return nil, err
	}

	// Initialize the TCP transport layer
	trans, err := raft.NewTCPTransport(s.config.RaftAdvertise.String(),
		s.config.RaftAdvertise, 3, s.config.RaftTimeout, s.config.LogOutput,
	)
	if err != nil {
		return nil, err
	}

	/// START TODO
	// For development purposes only. Add persistent storage layer
	var log raft.LogStore
	var stable raft.StableStore
	var snap raft.SnapshotStore

	store := raft.NewInmemStore()
	s.raftInmem = store
	stable = store
	log = store
	snap = raft.NewDiscardSnapshotStore()
	/// END TODO

	// Bootstrap the cluster if this is the only server and does not have
	// existing state.
	if s.isSingleServerCluster() {
		hasState, err := raft.HasExistingState(log, stable, snap)
		if err != nil {
			return nil, err
		}

		if !hasState {
			configuration := raft.Configuration{
				Servers: []raft.Server{
					{
						ID:      config.LocalID,
						Address: trans.LocalAddr(),
					},
				},
			}

			if err := raft.BootstrapCluster(config, log, stable, snap, trans, configuration); err != nil {
				return nil, err
			}
		}
	}

	return raft.NewRaft(config, s.fsm, log, stable, snap, trans)
}

// setupSerf sets up and initializes the Serf node.
func (s *Server) setupSerf() (*serf.Serf, error) {
	c := s.config.SerfConfig
	c.Init()

	// Tag the Serf member
	c.Tags["role"] = "openstate"
	c.Tags["addr"] = s.config.RaftAdvertise.String()
	c.Tags["id"] = s.config.NodeID

	// Setup logging
	logger := s.logger.StandardLogger(&log.StandardLoggerOptions{InferLevels: true})
	c.Logger = logger
	c.LogOutput = nil
	c.MemberlistConfig.Logger = logger
	c.MemberlistConfig.LogOutput = nil

	// General configurations
	c.NodeName = s.config.NodeName
	c.EventCh = s.eventCh
	c.RejoinAfterLeave = true
	c.LeavePropagateDelay = 1 * time.Second
	c.EnableNameConflictResolution = false
	c.MemberlistConfig.BindPort = s.config.SerfAdvertise.Port
	c.MemberlistConfig.AdvertisePort = s.config.SerfAdvertise.Port

	return serf.Create(c)
}

// bootstrapHandler joins the OpenState gossip ring given a list of peers.
//
// TODO
// Providing the list of peers via server config is a workaround for service
// discovery. We should instead be pulling the peer list from Consul or similar
// service.
func (s *Server) bootstrapHandler(peers []string) error {
	n, err := s.Join(peers)
	if err != nil {
		return fmt.Errorf("Contacted %d Nomad Servers: %v", n, err)
	}
	return nil
}

// IsLeader returns true if the server's raft node is the leader, otherwise false
func (s *Server) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

// isSingleServerCluster returns true if the expected number of bootstrapped
// servers is 1, otherwise false.
func (s *Server) isSingleServerCluster() bool {
	return s.config.BootstrapExpect == 1
}

// Join join's the server to the OpenSteate gossip ring. The target address(es)
// should be another node listening on the Serf address.
func (s *Server) Join(addrs []string) (int, error) {
	return s.serf.Join(addrs, true)
}

// TODO
// Run holds the server routine open. This should be replaced with the endpoint
// listener.
func (s *Server) Run() {
	for {
	}
}
