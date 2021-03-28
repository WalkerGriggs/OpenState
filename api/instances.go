package api

import (
	"fmt"
	"net/url"

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
