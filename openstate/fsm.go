package openstate

import (
	"encoding/json"
	"io"

	log "github.com/hashicorp/go-hclog"

	"github.com/hashicorp/raft"
)

type FSMConfig struct {
	// Logger is the logger used by the FSM
	Logger log.Logger
}

type FSM struct {
	// names are the hello-world state to replicate across the cluster
	names []string

	// Logger is the logger used by the FSM
	logger log.Logger
}

func NewFSM(config *FSMConfig) (*FSM, error) {
	return &FSM{
		names:  []string{},
		logger: config.Logger,
	}, nil
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	buf := log.Data
	msgType := MessageType(buf[0])

	switch msgType {
	case NameAddRequestType:
		return f.applyAddName(msgType, buf[1:], log.Index)
	}

	panic("Failed to apply log!")
}

func (f *FSM) applyAddName(reqType MessageType, buf []byte, index uint64) interface{} {
	var req NameAddRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	f.names = append(f.names, req.Name)

	return nil
}

// TODO
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil
}

// TODO
func (f *FSM) Restore(old io.ReadCloser) error {
	return nil
}
