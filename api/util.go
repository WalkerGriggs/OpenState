package api

import (
	"github.com/walkergriggs/openstate/fsm"
)

// Ftof converts an api.FSM to an fsm.FSM
func Ftof(m *FSM) (*fsm.FSM, error) {
	events := make([]*fsm.Event, len(m.Events))

	for i, event := range m.Events {
		e := &fsm.Event{
			Name: event.Name,
			Dst:  event.Dst,
			Src:  event.Src,
		}

		events[i] = e
	}

	return fsm.NewFSM(&fsm.FSMConfig{}, m.Initial, events)
}
