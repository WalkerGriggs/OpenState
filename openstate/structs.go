package openstate

import (
	"github.com/walkergriggs/openstate/api"
)

type MessageType uint8

const (
	TaskDefineRequestType MessageType = 0
)

type TaskDefineRequest struct {
	Task *api.Task
}
