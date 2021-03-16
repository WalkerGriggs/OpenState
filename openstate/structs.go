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

type TaskDefineRequest struct {
	Definition *api.Definition
}

type TaskRunRequest struct {
	Instance *Instance
}

// Definition is used to define task workflows.
type Definition struct {
	// Name is used to describe the task and all versions,
	Name string

	// Attributes are opaque descriptors to decorate the task.
	Attributes map[string]string

	// FSM is the API's serialized description of a desired finite state machine.
	// This FSM has no functionality aside from being used to create an instance's
	// fsm.FSM.
	FSM *api.FSM
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
	if d.Name == "" {
		return fmt.Errorf("Missing name")
	}

	if d.FSM == nil {
		return fmt.Errorf("Missing FSM")
	}

	return nil
}
