package openstate

import (
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

func (s *Server) monitorLeadership() {
	var weAreLeaderCh chan struct{}
	var leaderLoop sync.WaitGroup

	leaderCh := s.raft.LeaderCh()

	leaderStep := func(isLeader bool) {
		if isLeader {
			if weAreLeaderCh != nil {
				s.logger.Error("attempting to start the leader loop while running")
				return
			}

			weAreLeaderCh = make(chan struct{})
			leaderLoop.Add(1)
			go func(ch chan struct{}) {
				defer leaderLoop.Done()
				s.leaderLoop(ch)
			}(weAreLeaderCh)

			return
		}

		if weAreLeaderCh == nil {
			s.logger.Error("attempted to stop the leader loop while not running")
			return
		}

		close(weAreLeaderCh)
		leaderLoop.Wait()
		weAreLeaderCh = nil
		s.logger.Info("cluster leadership lost")
	}

	for {
		select {
		case isLeader := <-leaderCh:
			leaderStep(isLeader)
		}
	}
}

func (s *Server) leaderLoop(stopCh chan struct{}) {
	var reconcileCh chan serf.Member

RECONCILE:
	reconcileCh = nil
	interval := time.After(60 * time.Second)

	barrier := s.raft.Barrier(1 * time.Minute)
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

func (s *Server) reconcile() error {
	members := s.serf.Members()
	for _, member := range members {
		if err := s.reconcileMember(member); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) reconcileMember(member serf.Member) error {
	var err error

	switch member.Status {
	case serf.StatusAlive:
		err = s.addRaftPeer(member)
	case serf.StatusLeft, serf.StatusFailed:
		err = s.removeRaftPeer(member)
	}

	if err != nil {
		s.logger.Error("Failed to reconcile member", "member", member, "error", err)
		return err
	}
	return nil
}

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
