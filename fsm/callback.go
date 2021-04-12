package fsm

import (
	"context"
	"time"
)

type CallbackState struct {
	Status     string
	Running    bool
	Paused     bool
	Error      error
	StartedAt  time.Time
	FinishedAt time.Time
}

type Callback interface {
	Run(context.Context) (*CallbackState, error)
	State(context.Context) (*CallbackState, error)
	Wait(context.Context) (*CallbackState, error)
}
