package openstate

import (
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

type Task struct {
	Name string
	Tags []string
	FSM  *fsm.FSM
}
