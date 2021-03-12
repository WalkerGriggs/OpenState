package api

import (
	"github.com/walkergriggs/openstate/fsm"
)

// Tasks wraps the client and is used for task-specific endpoints
type Tasks struct {
	client *Client
}

// Task is used to serialize tasks
type Task struct {
	// Name is a globally unique name for the task.
	Name string `yaml:"name"`

	// Tags is a list of task-relevent attributes.
	Tags []string `yaml:"tags"`

	FSM *FSM `yaml:"state_machine"`
}

// FSM is used to serialize state machines
type FSM struct {
	// Initial is the state machines starting state
	Initial string `yaml:"initial_state"`

	// Events is a list of event descriptions.
	Events []*Event `yaml:"events"`
}

// Event is used to serialize state machine events
type Event struct {
	// Name is the event identifier and should be unique acrosss the state machine
	Name string `yaml:"name"`

	// Dst is the state the FSM will transition to after performing the event.
	Dst string `yaml:"destination"`

	// Src is a list of all possible source state. Each state should be unique, so
	// ideally this slice would be a set.
	Src []string `yaml:"sources"`
}

// FSMs wraps the client for task-specific endpoints
func (c *Client) Tasks() *Tasks {
	return &Tasks{client: c}
}

type (
	// TaskDefineRequest is used to serialize a Define request
	TaskDefineRequest struct {
		Task *Task
	}

	// TaskDefineResponse is used to serialize a Define response
	TaskDefineResponse struct {
		Index uint64
		Name  string
		Tags  []string
	}
)

// Define is used to create a new task.
func (t *Tasks) Define(task *Task) (*TaskDefineResponse, error) {
	req := &TaskDefineRequest{
		Task: task,
	}

	var res TaskDefineResponse
	if err := t.client.write("/v1/tasks", req, &res, nil); err != nil {
		return nil, err
	}

	return &res, nil
}

type (
	// TaskListRequest is used to serialize a List request
	TaskListRequest struct{}

	// TaskListResponse is used to serialize a List response
	TaskListResponse struct {
		Len   int
		Names []string
	}
)

// List is used to list all defined tasks.
func (t *Tasks) List() (*TaskListResponse, error) {
	var res TaskListResponse
	if err := t.client.query("/v1/tasks", &res, nil); err != nil {
		return nil, err
	}

	return &res, nil
}

// Ftof converts an api.FSM to an fsm.FSM
func (m *FSM) Ftof() (*fsm.FSM, error) {
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
