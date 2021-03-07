package openstate

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
)

func (s *Server) raftApply(t MessageType, msg interface{}) (interface{}, uint64, error) {
	future, err := s.raftApplyFuture(t, msg)
	if err != nil {
		return nil, 0, err
	}

	if err := future.Error(); err != nil {
		return nil, 0, err
	}

	return future.Response(), future.Index(), nil
}

func (s *Server) raftApplyFuture(t MessageType, msg interface{}) (raft.ApplyFuture, error) {
	buf, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("Failed to encode request: %v", err)
	}

	// Prepend the byte array with the MessageType
	buf = append([]byte{uint8(t)}, buf...)

	future := s.raft.Apply(buf, 30*time.Second)
	return future, nil
}
