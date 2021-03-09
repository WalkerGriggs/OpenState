package openstate

import (
	"encoding/json"
	"io"

	log "github.com/hashicorp/go-hclog"

	"github.com/hashicorp/raft"
)

type fsm struct {
	// names are the hello-world state to replicate across the cluster
	names []string

	// Logger is the logger used by the FSM
	logger log.Logger
}

type fsmConfig struct {
	// Logger is the logger used by the FSM
	Logger log.Logger
}

type fsmSnapshot struct {
	state []string
}

func NewFSM(config *fsmConfig) (*fsm, error) {
	return &fsm{
		names:  []string{},
		logger: config.Logger,
	}, nil
}

// Apply is invoked once a log entry is committed and persists the log to the
// FSm
func (f *fsm) Apply(log *raft.Log) interface{} {
	buf := log.Data
	msgType := MessageType(buf[0])

	switch msgType {
	case NameAddRequestType:
		return f.applyAddName(msgType, buf[1:], log.Index)
	}

	panic("Failed to apply log!")
}

func (f *fsm) applyAddName(reqType MessageType, buf []byte, index uint64) interface{} {
	var req NameAddRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	f.names = append(f.names, req.Name)

	return nil
}

// Snapshot supports log compaction. This call should return an FSMSnapshot
// which can be used to save a point-in-time snapshot of the FSM.
func (f *fsm) Snapshot() (raft.FSMSnapshot, error) {
	state := make([]string, len(f.names))
	copy(state, f.names)

	return &fsmSnapshot{state}, nil
}


// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous state.
func (f *fsm) Restore(rc io.ReadCloser) error {
	state := make([]string, 0)
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}

	f.names = state
	return nil
}

// Persist dumps all necessary state to the WriteCloser 'sink'.
func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(s.state)
		if err != nil {
			return err
		}

		if _, err := sink.Write(b); err != nil {
			return err
		}

		return sink.Close()
	}()

	// Cancel the sink if we failed to write it
	if err != nil {
		sink.Cancel()
	}

	return err
}

// Release is invoked when we are finished with the snapshot.
func (s *fsmSnapshot) Release() {}
