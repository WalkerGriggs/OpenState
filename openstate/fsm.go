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
	logger      log.Logger
	definitions map[string]*Definition
	instances   map[string]*Instance
}

// openstateFSMConfig is used to configure the openstateFSM
type openstateFSMConfig struct {
	Logger log.Logger
}

// openstateSnapshot implements raft.FSMSnapshot and is used to persist a
// point-in-time replica of the FSM's state to disk.
type openstateSnapshot struct {
	definitions map[string]*Definition
}

// NewFSM returns a FSM given a config.
func NewFSM(config *openstateFSMConfig) (*openstateFSM, error) {
	return &openstateFSM{
		definitions: make(map[string]*Definition),
		instances:   make(map[string]*Instance),
		logger:      config.Logger,
	}, nil
}

// Apply is invoked once a log entry is committed and persists the log to the
// FSM
func (f *openstateFSM) Apply(log *raft.Log) interface{} {
	buf := log.Data
	msgType := MessageType(buf[0])

	switch msgType {
	case TaskDefineRequestType:
		return f.applyDefineTask(msgType, buf[1:], log.Index)
	case TaskRunRequestType:
		return f.applyRunTask(msgType, buf[1:], log.Index)
	}

	panic("Failed to apply log!")
}

// applyDefineTask parses the task definition from the request and applies it to
// the Raft cluster
func (f *openstateFSM) applyDefineTask(reqType MessageType, buf []byte, index uint64) interface{} {
	var req TaskDefineRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	def := &Definition{
		Name:       req.Definition.Metadata.Name,
		Attributes: req.Definition.Metadata.Attributes,
		FSM:        req.Definition.FSM,
	}

	if err := def.Validate(); err != nil {
		return err
	}

	f.definitions[def.Name] = def
	return nil
}

// applyRunTask parses the task instance from the request and applies it to the
// Raft cluster
func (f *openstateFSM) applyRunTask(reqType MessageType, buf []byte, index uint64) interface{} {
	var req TaskRunRequest
	if err := json.Unmarshal(buf, &req); err != nil {
		f.logger.Error("decode raft log err %v", err)
		return err
	}

	f.instances[req.Instance.ID] = req.Instance
	return nil
}

// Snapshot supports log compaction. This call should return an FSMSnapshot
// which can be used to save a point-in-time snapshot of the FSM.
// TODO: Snapshot running instances
func (f *openstateFSM) Snapshot() (raft.FSMSnapshot, error) {
	defs := make(map[string]*Definition)
	for k, v := range f.definitions {
		defs[k] = v
	}

	return &openstateSnapshot{defs}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous state.
// TODO restore running instances
func (f *openstateFSM) Restore(rc io.ReadCloser) error {
	defs := make(map[string]*Definition)
	if err := json.NewDecoder(rc).Decode(&defs); err != nil {
		return err
	}

	f.definitions = defs
	return nil
}

// Persist dumps all necessary state to the WriteCloser 'sink'.
// TODO persist running instances
func (s *openstateSnapshot) Persist(sink raft.SnapshotSink) error {
	err := func() error {
		b, err := json.Marshal(s.definitions)
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
