package openstate

import (
	"encoding/json"
	"io"

	log "github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"

	"github.com/walkergriggs/openstate/openstate/state"
	"github.com/walkergriggs/openstate/openstate/structs"
)

// openstateFSM implements raft.FSM and is used for strongly consistent state
// replication across the cluster.
type openstateFSM struct {
	logger log.Logger

	// New way of doing things
	state *state.StateStore
}

// openstateFSMConfig is used to configure the openstateFSM
type openstateFSMConfig struct {
	Logger log.Logger
}

// openstateSnapshot implements raft.FSMSnapshot and is used to persist a
// point-in-time replica of the FSM's state to disk.
type openstateSnapshot struct {
	*state.StateSnapshot
}

// NewFSM returns a FSM given a config.
func NewFSM(config *openstateFSMConfig) (*openstateFSM, error) {
	fsm := &openstateFSM{
		logger: config.Logger,
	}

	var err error

	fsm.state, err = state.NewStateStore(&state.Config{})
	if err != nil {
		return nil, err
	}

	return fsm, nil
}

// Apply is invoked once a log entry is committed and persists the log to the
// FSM
func (f *openstateFSM) Apply(log *raft.Log) interface{} {
	buf := log.Data
	msgType := structs.MessageType(buf[0])

	switch msgType {
	case structs.TaskDefineRequestType:
		return f.applyDefineTask(msgType, buf[1:], log.Index)
	case structs.TaskRunRequestType:
		return f.applyRunTask(msgType, buf[1:], log.Index)
	}

	panic("Failed to apply log!")
}

// applyDefineTask parses the task definition from the request and applies it to
// the Raft cluster
func (f *openstateFSM) applyDefineTask(reqType structs.MessageType, buf []byte, index uint64) interface{} {
	var req structs.TaskDefineRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	def := req.Definition

	if err := def.Validate(); err != nil {
		return err
	}

	return f.state.InsertDefinition(def)
}

// applyRunTask parses the task instance from the request and applies it to the
// Raft cluster
func (f *openstateFSM) applyRunTask(reqType structs.MessageType, buf []byte, index uint64) interface{} {
	var req structs.TaskRunRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	return f.state.InsertInstance(req.Instance)
}

// Snapshot supports log compaction. This call should return an FSMSnapshot
// which can be used to save a point-in-time snapshot of the FSM.
func (f *openstateFSM) Snapshot() (raft.FSMSnapshot, error) {
	snap, err := f.state.Snapshot()
	if err != nil {
		return nil, err
	}

	return &openstateSnapshot{snap}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous state.
// TODO restore running instances
func (f *openstateFSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	if err := f.restoreDefinitions(rc); err != nil {
		return err
	}

	if err := f.restoreInstances(rc); err != nil {
		return err
	}

	return nil
}

func (f *openstateFSM) restoreDefinitions(rc io.ReadCloser) error {
	defs := make([]*structs.Definition, 0)
	if err := json.NewDecoder(rc).Decode(&defs); err != nil {
		return err
	}

	for _, def := range defs {
		f.state.InsertDefinition(def)
	}

	return nil
}

func (f *openstateFSM) restoreInstances(rc io.ReadCloser) error {
	instances := make([]*structs.Instance, 0)
	if err := json.NewDecoder(rc).Decode(&instances); err != nil {
		return err
	}

	for _, instance := range instances {
		f.state.InsertInstance(instance)
	}

	return nil
}

// Persist dumps all necessary state to the WriteCloser 'sink'.
// TODO persist running instances
func (s *openstateSnapshot) Persist(sink raft.SnapshotSink) error {
	if err := s.persistDefinitions(sink); err != nil {
		return err
	}

	if err := s.persistInstances(sink); err != nil {
		return err
	}

	return nil
}

func (s *openstateSnapshot) persistDefinitions(sink raft.SnapshotSink) error {
	err := func() error {
		defs, err := s.GetDefinitions()
		if err != nil {
			return err
		}

		b, err := json.Marshal(defs)
		if err != nil {
			return err
		}

		sink.Write(b)

		return nil
	}()

	if err != nil {
		sink.Cancel()
	}
	return err
}

func (s *openstateSnapshot) persistInstances(sink raft.SnapshotSink) error {
	err := func() error {
		instances, err := s.GetInstances()
		if err != nil {
			return err
		}

		b, err := json.Marshal(instances)
		if err != nil {
			return err
		}

		sink.Write(b)

		return nil
	}()

	if err != nil {
		sink.Cancel()
	}
	return err
}

// Release is invoked when we are finished with the snapshot.
func (s *openstateSnapshot) Release() {}
