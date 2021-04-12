package api

import (
	"os"

	"github.com/walkergriggs/openstate/driver/docker"
	"github.com/walkergriggs/openstate/fsm"
)

// Ftof converts an api.FSM to an fsm.FSM
func Ftof(m *FSM) (*fsm.FSM, error) {
	events := make([]*fsm.Event, len(m.Events))
	callbacks := make(map[string]fsm.Callback)

	for i, event := range m.Events {
		e := &fsm.Event{
			Name: event.Name,
			Dst:  event.Dst,
			Src:  event.Src,
		}

		if event.Callback != nil {
			callback, err := docker.NewCallback(&docker.CallbackConfig{
				Name:   "foo",
				Image:  event.Callback.Image,
				Writer: os.Stdout,
			})
			if err != nil {
				return nil, err
			}

			callbacks[event.Name] = callback
		}

		events[i] = e
	}

	return fsm.NewFSM(&fsm.FSMConfig{}, m.Initial, events, callbacks)
}
