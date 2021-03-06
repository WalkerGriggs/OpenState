package openstate

import (
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

// monitorLeadership listens on the server's Raft leader channel, and assumes
// leader responsibilities if neceesary. This method runs on each server, so
// server is responsible for reliquishing these responsibilities when they are
// no longer the leader.
func (s *Server) monitorLeadership() {
	var serverLeaderCh chan struct{}
	var leaderLoop sync.WaitGroup

	// If we are still the leader and the leaderLoop is not already running,
	// leaderStep adds a leaderLoop routine to the waitgroup and returns.
	//
	// If we are not the leader and the leaderLoop is still running, leaderStep
	leaderStep := func(isLeader bool) {
		if isLeader {
			if serverLeaderCh != nil {
				return
			}

			serverLeaderCh = make(chan struct{})
			leaderLoop.Add(1)
			go func(ch chan struct{}) {
				defer leaderLoop.Done()
				s.leaderLoop(ch)
			}(serverLeaderCh)

			return
		}

		if serverLeaderCh == nil {
			s.logger.Error("Attempted to stop the leader loop while not running")
			return
		}

		close(serverLeaderCh)
		leaderLoop.Wait()
		serverLeaderCh = nil
		s.logger.Info("cluster leadership lost")
	}

	// Listen over the server's leader channel. If we are the leader, call
	// leaderStep.
	raftLeaderCh := s.raft.LeaderCh()

	for {
		select {
		case isLeader := <-raftLeaderCh:
			leaderStep(isLeader)
		}
	}
}

// leaderLoop handles additional Raft leader responsibilities. Namely, it
// syncs Raft and Serf membership either 1) every minute 2) on Serf membership event.
func (s *Server) leaderLoop(stopCh chan struct{}) {
	var reconcileCh chan serf.Member

RECONCILE:
	reconcileCh = nil
	interval := time.After(60 * time.Second)

	// Barrier issues a command that blocks until all preceeding operations have
	// been applied to the FSM. It ensures the FSM reflects all queued writes.
	barrier := s.raft.Barrier(60 * time.Second)
	if err := barrier.Error(); err != nil {
		s.logger.Error("Failed to wait for barrier %v\n", err)
		goto WAIT
	}

	if err := s.reconcile(); err != nil {
		s.logger.Error("Failed to reconcile: %v\n", err)
		goto WAIT
	}

	reconcileCh = s.reconcileCh

WAIT:
	for {
		select {
		case <-stopCh:
			return
		case <-interval:
			goto RECONCILE
		case member := <-reconcileCh:
			s.reconcileMember(member)
		}
	}
}

// reconcile is used to sync members from Serf to Raft.
func (s *Server) reconcile() error {
	members := s.serf.Members()
	for _, member := range members {
		if err := s.reconcileMember(member); err != nil {
			return err
		}
	}
	return nil
}

// reconcile is used to sync a specific member from Serf to Raft.
func (s *Server) reconcileMember(member serf.Member) error {
	var err error

	switch member.Status {
	case serf.StatusAlive:
		err = s.addRaftPeer(member)
	case serf.StatusLeft, serf.StatusFailed:
		err = s.removeRaftPeer(member)
	}

	if err != nil {
		s.logger.Error("Failed to reconcile member", "member", member.Name, "error", err)
		return err
	}
	return nil
}

// addRaft adds a Serf member as a Raft peer. This function is idempotent.
func (s *Server) addRaftPeer(m serf.Member) error {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	addr := m.Tags["addr"]
	nodeID := m.Tags["id"]

	for _, server := range configFuture.Configuration().Servers {
		if server.Address == raft.ServerAddress(addr) && server.ID == raft.ServerID(nodeID) {
			return nil
		}
		// TODO - handle if the server has a mismatched address or server ID
	}

	addFuture := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if err := addFuture.Error(); err != nil {
		s.logger.Error("Failed to add server to Raft cluster: %v", err)
		return err
	}
	return nil
}

// removeRaftPeer removes a Serf member from Raft peers. This function is idempotent.
func (s *Server) removeRaftPeer(m serf.Member) error {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	nodeID := m.Tags["id"]

	for _, server := range configFuture.Configuration().Servers {
		if server.ID == raft.ServerID(nodeID) {
			s.logger.Info("Removing server", "member", m)

			removeFuture := s.raft.RemoveServer(raft.ServerID(nodeID), 0, 0)
			if err := removeFuture.Error(); err != nil {
				s.logger.Error("Failed to reomve server from Raft cluster: %v", err)
				return err
			}
		}
	}
	return nil
}
