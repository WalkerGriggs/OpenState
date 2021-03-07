package openstate

import (
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

func (s *Server) serfEventHandler() {
	for {
		select {
		case e := <-s.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				s.memberJoin(e.(serf.MemberEvent))
			case serf.EventMemberLeave:
				s.memberLeave(e.(serf.MemberEvent))
			case serf.EventMemberFailed:
				s.memberFailed(e.(serf.MemberEvent))
			default:
				s.logger.Warn("unhandled serf event")
			}
		}
	}
}

// memberJoin handles serf.EventMemberJoin.
func (s *Server) memberJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		parts, ok := isServer(m)
		if !ok {
			s.logger.Warn("Non-server in gossip ring.")
			continue
		}

		// TODO: check if the server is a known peer

		s.logger.Info("Adding peer", "peer", m.Name)
		s.peers[raft.ServerAddress(parts.raft_addr.String())] = parts
	}

	s.memberEvent(me)
}

// memberLeave handles serf.EventMemberLeave. Functionally equivalent to
// memberFailed
func (s *Server) memberLeave(me serf.MemberEvent) {
	s.memberFailed(me)
}

// memberFailed handles serf.EventMemberFailed.
func (s *Server) memberFailed(me serf.MemberEvent) {
	for _, m := range me.Members {
		parts, ok := isServer(m)
		if !ok {
			s.logger.Warn("Non-server in gossip ring.")
			continue
		}

		s.logger.Info("Removing peer", "peer", m.Name)
		delete(s.peers, raft.ServerAddress(parts.raft_addr.String()))

		// Is this the best way to handle a Serf node heading offline without
		// service discovery?
		if m.Status == serf.StatusFailed {
			if err := s.serf.RemoveFailedNode(m.Name); err != nil {
				s.logger.Error("Failed to remove failed Serf member", "member", m.Name, "error", err)
			}
		}
	}

	s.memberEvent(me)
}

// memberEvent pushes Serf members with a changed status over the server's
// reconcile channel. The leading Raft node will consume these members in
// leaderLoop and will adjust that node's Raft membership status accordingly.
func (s *Server) memberEvent(me serf.MemberEvent) {
	if !s.IsLeader() {
		return
	}

	for _, m := range me.Members {
		select {
		case s.reconcileCh <- m:
		default:
		}
	}
}
