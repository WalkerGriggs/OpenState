package api

import (
	"fmt"
	"net/url"
)

// Instances wraps the client and is used for instance-specific endpoints
type Instances struct {
	client *Client
}

// Instances wraps the client for instance-specific endpoints
func (c *Client) Instances() *Instances {
	return &Instances{client: c}
}

type (
	// InstanceRunRequest is used to serialize an Event request
	InstanceEventRequest struct {
		EventName string
	}

	// InstanceEventResponse is used to serialize an Event response
	InstanceEventResponse struct {
		CurrentState string
	}
)

// Event is used to perform an event against a task instance
func (t *Instances) Event(instance, event string) (*InstanceEventResponse, error) {
	req := &InstanceEventRequest{
		EventName: event,
	}

	var res InstanceEventResponse
	err := t.client.write(fmt.Sprintf("/v1/instance/%s/event", url.PathEscape(instance)), req, &res, nil)
	if err != nil {
		return nil, err
	}

	return &res, nil
}
