package docker

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/walkergriggs/openstate/fsm"
)

type CallbackConfig struct {
	Name   string
	Image  string
	Writer io.Writer
}

type Callback struct {
	config *CallbackConfig
	engine *Engine
}

func (c *Callback) MarshalText() ([]byte, error) {
	b, err := json.Marshal(
		struct {
			Config *CallbackConfig `json:"config"`
		}{
			Config: c.config,
		},
	)

	return b, err
}

func (c *Callback) UnmarshalText(b []byte) error {
	callback := struct {
		Config *CallbackConfig `json:"config"`
	}{}

	err := json.Unmarshal(b, &callback)
	if err != nil {
		return err
	}

	c.config = callback.Config

	return nil
}

func NewCallback(config *CallbackConfig) (*Callback, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}

	return &Callback{
		config: config,
		engine: NewEngine(cli, Config{}),
	}, nil
}

func (c *Callback) Run(ctx context.Context) (*fsm.CallbackState, error) {
	containerConfig := &ContainerConfig{
		Image: c.config.Image,
		Name:  c.config.Name,
	}

	state, err := c.engine.Run(ctx, containerConfig, c.config.Writer)
	if err != nil {
		return nil, err
	}

	return stos(state), nil
}

func (c *Callback) State(ctx context.Context) (*fsm.CallbackState, error) {
	state, err := c.engine.state(ctx, c.config.Name)
	if err != nil {
		return nil, err
	}

	return stos(state), nil
}

func (c *Callback) Wait(ctx context.Context) (*fsm.CallbackState, error) {
	state, err := c.engine.wait(ctx, c.config.Name)
	if err != nil {
		return nil, err
	}

	return stos(state), nil
}

// stos (state to state) is used to translate the docker specific
// types.Container state to the general CallbackState.
func stos(state *types.ContainerState) *fsm.CallbackState {
	startedAt, _ := time.Parse(time.RFC3339, state.StartedAt)
	finishedAt, _ := time.Parse(time.RFC3339, state.FinishedAt)

	return &fsm.CallbackState{
		Status:     state.Status,
		Running:    state.Running,
		Paused:     state.Paused,
		Error:      errors.New(state.Error),
		StartedAt:  startedAt,
		FinishedAt: finishedAt,
	}
}
