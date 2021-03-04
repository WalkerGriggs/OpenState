package openstate

import (
	"fmt"

	"github.com/hashicorp/serf/serf"
)

func (s *Server) serfEventHandler() {
	for {
		select {
		case e := <-s.eventCh:
			switch e.EventType() {
			case serf.EventMemberJoin:
				s.nodeJoin(e.(serf.MemberEvent))
				s.localMemberEvent(e.(serf.MemberEvent))
			default:
				s.logger.Warn("unhandled serf event")
			}
		}
	}
}

// TODO: check if member is a valid OpenState server
//       see isNomadServer()
//
// TODO: check if the server is a known peer
//       if known, continue outer loop
func (s *Server) nodeJoin(me serf.MemberEvent) {
	for _, m := range me.Members {
		s.peers = append(s.peers, m.Name)
	}
}

func (s *Server) localMemberEvent(me serf.MemberEvent) {
	if !s.IsLeader() {
		return
	}

	isReap := me.EventType() == serf.EventMemberReap

	for _, m := range me.Members {
		if isReap {
			m.Status = serf.MemberStatus(-1)
		}

		select {
		case s.reconcileCh <- m:
		default:
		}
	}
}
