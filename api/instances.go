package api

import (
	"fmt"
	"net/url"
	"strings"
	"text/tabwriter"

	"github.com/walkergriggs/openstate/fsm"
)

// Instances wraps the client and is used for instance-specific endpoints
type Instances struct {
	client *Client
}

// Instances wraps the client for instance-specific endpoints
func (c *Client) Instances() *Instances {
	return &Instances{client: c}
}

// Instance is used to run a defined workflow.
type Instance struct {
	// ID is used to describe the specific instance. This should be globally
	// unique.
	ID string

	// Definition is used to referece the definition from which the instance was
	// created. The definition should be treated as an immutable constant, as it's
	// likely shared between multiple instances.
	Definition *Definition

	// FSM is the graph defined, event driven state machine which provides the
	// instances backing functionality.
	FSM *fsm.FSM
}

// Summarize is used to derrive an InstanceSummary from an Instance object.
func (i *Instance) Summarize() *InstanceSummary {
	return &InstanceSummary{
		ID:         i.ID,
		Definition: i.Definition.Name,
		Current:    i.FSM.State(),
	}
}

// InstanceSummary is used as a point-in-time summary or the instance object.
// The InstanceSummary doesn't expose any functionality itself, but is used to
// convey high-level information to clients.
type InstanceSummary struct {
	ID         string
	Definition string
	Current    string
}

// String is used to provide a string representation of an InstanceSummary. The
// key/values are columar and tab aligned.
func (i *InstanceSummary) String() string {
	builder := &strings.Builder{}
	writer := tabwriter.NewWriter(builder, 0, 0, 1, ' ', 0)

	fmt.Fprintf(writer, "ID\t = %s\n", i.ID)
	fmt.Fprintf(writer, "Definition\t = %s\n", i.Definition)
	fmt.Fprintf(writer, "Current State\t = %v\n", i.Current)
	writer.Flush()

	return builder.String()
}

type (
	// InstanceRunRequest is used to serialize an Event request
	InstanceEventRequest struct {
		EventName string
	}

	// InstanceEventResponse is used to serialize an Event response
	InstanceEventResponse struct {
		Instance *Instance
	}
)

// Event is used to perform an event against a task instance
func (t *Instances) Event(instance, event string) (*InstanceEventResponse, error) {
	req := &InstanceEventRequest{
		EventName: event,
	}

	path := fmt.Sprintf("/v1/instance/%s/event", url.PathEscape(instance))

	var res InstanceEventResponse
	err := t.client.write(path, req, &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
