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

// FSMDefineRequest is used to define a new task
type TaskDefineRequest struct {
	Task *Task
}

// FSMDefineResponse is used to response to an task definition request.
type TaskDefineResponse struct {
}

type TaskListRequest struct {
}

type TaskListResponse struct {
	Len int
}

// FSMs wraps the client for task-specific endpoints
func (c *Client) Tasks() *Tasks {
	return &Tasks{client: c}
}

//List is used to list all defined tasks.
func (t *Tasks) List() (*TaskListResponse, error) {
	var res TaskListResponse
	if err := t.client.query("/v1/tasks", &res, nil); err != nil {
		return nil, err
	}

	return &res, nil
}

// Define is used to create a new task.
func (t *Tasks) Define(task *Task) (*TaskDefineResponse, error) {
	req := &TaskDefineRequest{
		Task: task,
	}

	if err := t.client.write("/v1/tasks", req, nil, nil); err != nil {
		return nil, err
	}

	return nil, nil
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
