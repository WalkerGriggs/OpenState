package mocks

import (
	"github.com/walkergriggs/openstate/api"
	"github.com/walkergriggs/openstate/fsm"
	"github.com/walkergriggs/openstate/openstate/structs"
)

func Definition() *structs.Definition {
	return &structs.Definition{
		Name: "mock-def",
		Attributes: map[string]string{
			"author": "mock-author",
		},
		FSM: apiFSM(),
	}
}

func Instance() *structs.Instance {
	fsm, _ := api.Ftof(apiFSM())

	return &structs.Instance{
		ID:         "mock-def-1234",
		Definition: Definition(),
		FSM:        fsm,
	}
}

func FSM() *fsm.FSM {
	config := &fsm.FSMConfig{}
	initial := "green"
	events := []*fsm.Event{
		&fsm.Event{"turn_green", "green", []string{"red"}},
		&fsm.Event{"turn_yellow", "yellow", []string{"green"}},
		&fsm.Event{"turn_red", "red", []string{"yellow"}},
	}

	fsm, _ := fsm.NewFSM(config, initial, events)

	return fsm
}

func apiFSM() *api.FSM {
	initial := "green"
	events := []*api.Event{
		&api.Event{"turn_green", "green", []string{"red"}},
		&api.Event{"turn_yellow", "yellow", []string{"green"}},
		&api.Event{"turn_red", "red", []string{"yellow"}},
	}

	return &api.FSM{
		Initial: initial,
		Events:  events,
	}
}
