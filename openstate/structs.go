package openstate

import (
	"fmt"

	"github.com/walkergriggs/openstate/api"
	"github.com/walkergriggs/openstate/fsm"
)

type MessageType uint8

const (
	TaskDefineRequestType MessageType = 0
)

type TaskDefineRequest struct {
	Task *api.Task
}

// Task is used to define and execute workflows.
type Task struct {
	// Name is used to describe the task and all versions
	Name string

	// Attributes are opaque descriptors to decorate the task.
	Attributes map[string]string

	// FSM is the graph defined, event driven state machines which provides
	// the Task's backing functionality.
	FSM *fsm.FSM
}

// Validate is used to check the validity of a task object. Validate is not
// considered comprehensive at this time.
func (t *Task) Validate() error {
	if t.Name == "" {
		return fmt.Errorf("Missing name")
	}

	if t.FSM == nil {
		return fmt.Errorf("Missing FSM")
	}

	return t.FSM.Validate()
}
