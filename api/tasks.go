package api

import (
	"fmt"
	"net/url"
)

// Tasks wraps the client and is used for task-specific endpoints
type Tasks struct {
	client *Client
}

// Definition is used to serialize task definitions
type Definition struct {
	// Metadata groups task-related descriptors
	Metadata *Metadata `yaml:"metadata"`

	FSM *FSM `yaml:"state_machine"`
}

// Metadata groups task-related descriptors
type Metadata struct {
	// Name is a globally unique name for the task.
	Name string `yaml:"name"`

	// Attributes is a map of task-relevant key-values pairs
	Attributes map[string]string `yaml:"attributes"`
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
		Definition *Definition
	}

	// TaskDefineResponse is used to serialize a Define response
	TaskDefineResponse struct {
		Index      uint64
		Name       string
		Attributes map[string]string
	}
)

// Define is used to create a new task.
func (t *Tasks) Define(def *Definition) (*TaskDefineResponse, error) {
	req := &TaskDefineRequest{
		Definition: def,
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

type (
	// TaskRunRequest is used to serialize a Run request
	TaskRunRequest struct{}

	// TaskRunResponse is used to serialize a Run resonse
	TaskRunResponse struct {
		InstanceID string
	}
)

// Run is used to initialize a running task instance from a definition
func (t *Tasks) Run(taskName string) (*TaskRunResponse, error) {
	req := &TaskRunRequest{}

	var res TaskRunResponse
	err := t.client.write(fmt.Sprintf("/v1/task/%s/run", url.PathEscape(taskName)), req, &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

type (
	// TaskPsRequest is used to serialize a Ps request
	TaskPsRequest struct{}

	// TaskPsResponse is used to serialize a Ps response
	TaskPsResponse struct {
		Len int
		IDs []string
	}
)

// Ps is used to list running task instances for a given task definition
func (t *Tasks) Ps(taskName string) (*TaskPsResponse, error) {
	var res TaskPsResponse
	err := t.client.query(fmt.Sprintf("/v1/task/%s/ps", url.PathEscape(taskName)), &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
