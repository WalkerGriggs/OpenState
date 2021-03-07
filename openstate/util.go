package openstate

import (
	"net"

	"github.com/hashicorp/serf/serf"
)

type serverParts struct {
	raft_addr *net.TCPAddr
	serf_addr *net.TCPAddr
	http_addr *net.TCPAddr
	role      string
	id        string
}

func isServer(m serf.Member) (*serverParts, bool) {
	if m.Tags["role"] != "openstate" {
		return nil, false
	}

	raft_addr, err := net.ResolveTCPAddr("tcp", m.Tags["raft_addr"])
	if err != nil {
		return nil, false
	}

	serf_addr, err := net.ResolveTCPAddr("tcp", m.Tags["serf_addr"])
	if err != nil {
		return nil, false
	}

	http_addr, err := net.ResolveTCPAddr("tcp", m.Tags["http_addr"])
	if err != nil {
		return nil, false
	}

	return &serverParts{
		raft_addr: raft_addr,
		serf_addr: serf_addr,
		http_addr: http_addr,
		role:      m.Tags["role"],
		id:        m.Tags["id"],
	}, true
}
