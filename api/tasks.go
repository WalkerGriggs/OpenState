package api

import (
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"
)

// Tasks wraps the client and is used for task-specific endpoints
type Tasks struct {
	client *Client
}

// FSMs wraps the client for task-specific endpoints
func (c *Client) Tasks() *Tasks {
	return &Tasks{client: c}
}

// Definition is used to serialize task definitions
type Definition struct {
	// Metadata groups task-related descriptors
	Metadata *DefinitionMetadata `yaml:"metadata"`

	// FSM is the blueprint for a graph defined, event driven state machine.
	// api.FSM doesn't directly expose any functionality, but it used to
	// initialize a fsm.FSM.
	FSM *FSM `yaml:"state_machine"`
}

// Summarize is used to derrive an DefinitionSummary from a Definition object.
func (d *Definition) Summarize() *DefinitionSummary {
	return &DefinitionSummary{
		Name:    d.Metadata.Name,
		Initial: d.FSM.Initial,
		Events:  d.FSM.EventNames(),
	}
}

// DefinitionSummary is used as a point-in-time summary or the definition object.
// The DefinitionSummary doesn't expose any functionality itself, but is used to
// convey high-level information to clients.
type DefinitionSummary struct {
	Name    string
	Initial string
	Events  []string
}

// String is used to provide a string representation of an DefinitionSummary.
// The key/values are columar and tab aligned.
func (d *DefinitionSummary) String() string {
	builder := &strings.Builder{}
	writer := tabwriter.NewWriter(builder, 0, 0, 1, ' ', 0)

	fmt.Fprintf(writer, "Name\t = %s\n", d.Name)
	fmt.Fprintf(writer, "Initial State\t = %s\n", d.Initial)
	fmt.Fprintf(writer, "Events\t = %v\n", d.Events)
	writer.Flush()

	return builder.String()
}

// Metadata groups task-related descriptors
type DefinitionMetadata struct {
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

// EventNames is used to parse the event list and return a list of just the
// event names.
func (m *FSM) EventNames() (events []string) {
	for _, event := range m.Events {
		events = append(events, event.Name)
	}
	return
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

type (
	// TaskDefineRequest is used to serialize a Define request
	TaskDefineRequest struct {
		Definition *Definition
	}

	// TaskDefineResponse is used to serialize a Define response
	TaskDefineResponse struct {
		Definition *Definition
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
		Definitions []*Definition
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
		Instance *Instance
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
		Instances []*Instance
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
