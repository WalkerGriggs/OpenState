package openstate

import (
	"fmt"
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
				fmt.Println("attempting to start the leader loop while running")
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
			fmt.Println("attempted to stop the leader loop while not running")
			return
		}

		fmt.Println("shutting down leader loop")
		close(weAreLeaderCh)
		leaderLoop.Wait()
		weAreLeaderCh = nil
		fmt.Println("cluster leadership lost")
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
		fmt.Printf("Failed to wait for barrier %v\n", err)
		goto WAIT
	}

	if err := s.reconcile(); err != nil {
		fmt.Printf("Failed to reconcile: %v\n", err)
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
	switch member.Status {

	case serf.StatusAlive:
		return s.addRaftPeer(member)
		// case serf.StatusLeft, StatusReap:
		//	err = s.removeRaftPeer(member, parts)
	default:
		return nil
	}
}

func (s *Server) addRaftPeer(m serf.Member) error {
	configFuture := s.raft.GetConfiguration()
	if err := configFuture.Error(); err != nil {
		return err
	}

	addr := m.Tags["addr"]
	nodeID := m.Tags["id"]

	for _, server := range configFuture.Configuration().Servers {
		if server.Address == raft.ServerAddress(addr) || server.ID == raft.ServerID(nodeID) {
			fmt.Println("Server has already joined the cluster")
			return nil
		}
		// TODO - handle if the server has a mismatched address or server ID
	}

	addFuture := s.raft.AddVoter(raft.ServerID(nodeID), raft.ServerAddress(addr), 0, 0)
	if err := addFuture.Error(); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
