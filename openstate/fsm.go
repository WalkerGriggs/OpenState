package openstate

import (
	"encoding/json"
	"io"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
)

// openstateFSM implements raft.FSM and is used for strongly consistent state
// replication across the cluster.
type openstateFSM struct {
	logger log.Logger

	// tasks are the state to replicate across the cluster
	tasks []*Task
}

// openstateFSMConfig is used to configure the openstateFSM
type openstateFSMConfig struct {
	Logger log.Logger
}

// openstateSnapshot implements raft.FSMSnapshot and is used to persist a
// point-in-time replica of the FSM's state to disk.
type openstateSnapshot struct {
	state []*Task
}

// NewFSM returns a FSM given a config.
func NewFSM(config *openstateFSMConfig) (*openstateFSM, error) {
	return &openstateFSM{
		tasks:  make([]*Task, 0),
		logger: config.Logger,
	}, nil
}

// Apply is invoked once a log entry is committed and persists the log to the
// FSm
func (f *openstateFSM) Apply(log *raft.Log) interface{} {
	buf := log.Data
	msgType := MessageType(buf[0])

	switch msgType {
	case TaskDefineRequestType:
		return f.applyDefineTask(msgType, buf[1:], log.Index)
	}

	panic("Failed to apply log!")
}

func (f *openstateFSM) applyDefineTask(reqType MessageType, buf []byte, index uint64) interface{} {
	var req TaskDefineRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	fsm, err := req.Task.FSM.Ftof()
	if err != nil {
		return err
	}

	task := &Task{
		Name: req.Task.Name,
		Tags: req.Task.Tags,
		FSM:  fsm,
	}

	f.tasks = append(f.tasks, task)

	return nil
}

// Snapshot supports log compaction. This call should return an FSMSnapshot
// which can be used to save a point-in-time snapshot of the FSM.
func (f *openstateFSM) Snapshot() (raft.FSMSnapshot, error) {
	state := make([]*Task, len(f.tasks))
	copy(state, f.tasks)

	return &openstateSnapshot{state}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous state.
func (f *openstateFSM) Restore(rc io.ReadCloser) error {
	state := make([]*Task, 0)
	if err := json.NewDecoder(rc).Decode(&state); err != nil {
		return err
	}

	f.tasks = state
	return nil
}

// Persist dumps all necessary state to the WriteCloser 'sink'.
func (s *openstateSnapshot) Persist(sink raft.SnapshotSink) error {
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

	if err != nil {
		sink.Cancel()
	}
	return err
}

// Release is invoked when we are finished with the snapshot.
func (s *openstateSnapshot) Release() {}
