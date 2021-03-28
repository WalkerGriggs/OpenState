package openstate

import (
	"fmt"

	"github.com/walkergriggs/openstate/api"
	"github.com/walkergriggs/openstate/fsm"
)

type MessageType uint8

const (
	TaskDefineRequestType MessageType = 0
	TaskRunRequestType    MessageType = 1
)

type (
	// TaskListRequest is used to serialize a List request
	TaskListRequest struct{}

	// TaskListResponse is used to serialize a List response
	TaskListResponse struct {
		Definitions []*Definition
	}
)

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

type (
	// TaskRunRequest is used to serialize a Run request
	TaskRunRequest struct {
		Instance *Instance
	}

	// TaskRunResponse is used to serialize a Run resonse
	TaskRunResponse struct {
		Instance *Instance
	}
)

type (
	// TaskPsRequest is used to serialize a Ps request
	TaskPsRequest struct{}

	// TaskPsResponse is used to serialize a Ps response
	TaskPsResponse struct {
		Instances []*Instance
	}
)

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

// Definition is used to define task workflows.
type Definition struct {
	// Metadata groups task-related descriptors
	Metadata *DefinitionMetadata

	// FSM is the API's serialized description of a desired finite state machine.
	// This FSM has no functionality aside from being used to create an instance's
	// fsm.FSM.
	FSM *api.FSM
}

// Metadata groups task-related descriptors
type DefinitionMetadata struct {
	// Name is used to describe the task and all versions,
	Name string

	// Attributes are opaque descriptors to decorate the task.
	Attributes map[string]string
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

// Validate is used to check the validity of a task object. Validate is not
// considered comprehensive at this time.
func (d *Definition) Validate() error {
	if d.Metadata.Name == "" {
		return fmt.Errorf("Missing name")
	}

	if d.FSM == nil {
		return fmt.Errorf("Missing FSM")
	}

	return nil
}
