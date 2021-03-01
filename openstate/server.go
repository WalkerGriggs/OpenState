package openstate

import (
	"fmt"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

type Server struct {
	// Server configuration with reasonable defaults.
	config *Config

	// eventCh is used to receive Serf events.
	eventCh chan serf.Event

	// Finite State Machine used to maintain state across Raft nodes
	fsm *FSM

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
		reconcileCh: make(chan serf.Member, 32),
		eventCh:     make(chan serf.Event, 256),
	}

	var err error

	if err := s.setupRaft(); err != nil {
		return nil, fmt.Errorf("Failed to start Raft: %v", err)
	}

	s.serf, err = s.setupSerf(c.SerfConfig)
	if err != nil {
		return nil, fmt.Errorf("Failed to start serf: %v", err)
	}

	// Handle Serf events
	go s.serfEventHandler()

	// Monitor as assume leadership responsibilities
	go s.monitorLeadership()

	// Bootstrap the new server
	go s.bootstrapHandler(c.Peers)

	return s, nil
}

func (s *Server) setupRaft() error {
	config := s.config.RaftConfig

	var err error
	s.fsm, err = NewFSM()
	if err != nil {
		return err
	}

	trans, err := raft.NewTCPTransport(s.config.RaftAdvertise.String(),
		s.config.RaftAdvertise,
		3,
		s.config.RaftTimeout,
		s.config.LogOutput,
	)
	if err != nil {
		return err
	}

	config.LocalID = raft.ServerID(s.config.NodeID)

	var log raft.LogStore
	var stable raft.StableStore
	var snap raft.SnapshotStore

	/// START TODO
	// For development purposes only. Add persistent storage layer
	store := raft.NewInmemStore()
	s.raftInmem = store
	stable = store
	log = store
	snap = raft.NewDiscardSnapshotStore()
	/// END TODO

	// If we are a single server cluster and the state is clean then we can
	// bootstrap now.
	if s.isSingleServerCluster() {
		hasState, err := raft.HasExistingState(log, stable, snap)
		if err != nil {
			return err
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
				return err
			}
		}
	}

	s.raft, err = raft.NewRaft(config, s.fsm, log, stable, snap, trans)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) setupSerf(c *serf.Config) (*serf.Serf, error) {
	c.Init()

	c.Tags["role"] = "openstate"
	c.Tags["addr"] = s.config.RaftAdvertise.String()
	c.Tags["id"] = s.config.NodeID

	c.NodeName = s.config.NodeName
	c.EventCh = s.eventCh

	c.LogOutput = c.LogOutput
	c.RejoinAfterLeave = true
	c.LeavePropagateDelay = 1 * time.Second
	c.EnableNameConflictResolution = false

	c.MemberlistConfig.LogOutput = c.LogOutput
	c.MemberlistConfig.BindPort = s.config.SerfAdvertise.Port
	c.MemberlistConfig.AdvertisePort = s.config.SerfAdvertise.Port

	return serf.Create(c)
}

func (s *Server) bootstrapHandler(peers []string) error {
	numServersContacted, err := s.Join(peers)
	if err != nil {
		return fmt.Errorf("contacted %d Nomad Servers: %v", numServersContacted, err)
	}
	return nil
}

func (s *Server) IsLeader() bool {
	return s.raft.State() == raft.Leader
}

func (s *Server) isSingleServerCluster() bool {
	return s.config.BootstrapExpect == 1
}

func (s *Server) Join(addrs []string) (int, error) {
	return s.serf.Join(addrs, true)
}

func (s *Server) Run() {
	for {
	}
}
