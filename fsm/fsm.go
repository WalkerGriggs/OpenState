package fsm

import (
	"context"
	"fmt"
	"sync"
)

// FSMConfig is a no-op for now.
type FSMConfig struct{}

// FSM is the core Finite State Machine implementation. And follows a strict,
// gkaph-defined architecture where are transitions are event-based.
// TODO Implement callbacks for a given driver interface. Callbacks sound be
//      more than function pointers.
type FSM struct {
	// config is the FSM config struct
	config *FSMConfig

	// current is the current state of the FSM. This state should not be
	// manipulated directly.
	current string

	// stateMu is a read-write mutex lock to gaurd against state-related
	// race conditions.
	stateMu sync.RWMutex

	// transitions is a map of all unique transitions, where the keys are
	// built from unique event name/src pairs.
	transitions map[eParts]string

	// callbacks is a map of all unique callbacks, where they key is the
	// event name
	callbacks map[string]Callback
}

// Event describes a single event, its origin states, and its intended
// destination state
type Event struct {
	// Name is the event identifier and should be unique acrosss the state machine
	Name string

	// Dst is the state the FSM will transition to after performing the event.
	Dst string

	// Src is a list of all possible source state. Each state should be unique, so
	// ideally this slice would be a set.
	Src []string
}

type Events []*Event

// eParts describes the component parts of specific event/src pair. It is used
// as a unique key to map FSM transitions.
type eParts struct {
	event string
	src   string
}

// NewFSM create a FSM given an initial state and a list of events.
// TODO raise an error if the initial state is not a valid src state of any
//      event.
func NewFSM(config *FSMConfig, initial string, events Events, callbacks map[string]Callback) (*FSM, error) {
	fsm := &FSM{
		current:     initial,
		config:      config,
		callbacks:   callbacks,
		transitions: make(map[eParts]string),
	}

	for _, event := range events {
		for _, src := range event.Src {
			fsm.transitions[eParts{event.Name, src}] = event.Dst
		}
	}

	return fsm, nil
}

// State returns the FSM's current state.
func (fsm *FSM) State() string {
	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()
	return fsm.current
}

func (fsm *FSM) Callbacks() map[string]Callback {
	return fsm.callbacks
}

// Can checks if the state machine can perform the given event considering the
// FSM's current state.
func (fsm *FSM) Can(event string) bool {
	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()
	_, ok := fsm.transitions[eParts{event, fsm.current}]
	return ok
}

// Cannot is the inverse of Can
func (fsm *FSM) Cannot(event string) bool {
	return !fsm.Can(event)
}

// Do runs a given event against the state machine. It first checks if the FSM
// is in an allowed state before transitioning.
func (fsm *FSM) Do(event string) error {
	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()

	dst, ok := fsm.transitions[eParts{event, fsm.current}]
	if !ok {
		return fmt.Errorf("FSM cannot %s", event)
	}

	// TODO Add a lock
	// TODO Wait for callback to finish? How do we want to handle long running queries?
	callback := fsm.callbacks[event]
	if _, err := callback.Run(context.Background()); err != nil {
		return err
	}

	fsm.stateMu.RUnlock()
	defer fsm.stateMu.RLock()
	fsm.transition(dst)
	return nil
}

// AvailableTransitions lists all available events given for the current state.
func (fsm *FSM) AvailableEvents() (trans []string) {
	fsm.stateMu.RLock()
	defer fsm.stateMu.RUnlock()

	for key := range fsm.transitions {
		if key.src == fsm.current {
			trans = append(trans, key.event)
		}
	}
	return
}

// transition transitions the state machine to the destination state.
func (fsm *FSM) transition(dst string) {
	fsm.stateMu.Lock()
	defer fsm.stateMu.Unlock()
	fsm.current = dst
}

// Edges returns a list of all unique edges between source and destination
// states.
func (m *FSM) Edges() map[string][]string {
	edges := make(map[string][]string)

	for eparts, dst := range m.transitions {
		if _, ok := edges[eparts.src]; !ok {
			edges[eparts.src] = []string{}
		}

		edges[eparts.src] = append(edges[eparts.src], dst)
	}

	return edges
}

// Validate is used to check the validity of the FSM. Validate is not considered
// comprehensive at this time.
func (m *FSM) Validate() error {
	return nil
}
