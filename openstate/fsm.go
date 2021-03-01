package openstate

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/hashicorp/raft"
)

type FSM struct {
	names []string
}

func NewFSM() (*FSM, error) {
	return &FSM{
		names: []string{},
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
		fmt.Println("decode raft log err %v", err)
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
