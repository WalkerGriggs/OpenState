package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"

	// "github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type Config struct{}

type Engine struct {
	config Config
	client client.APIClient
}

type ContainerConfig struct {
	Name  string
	Image string
	// WorkingDir string
	// Entrypoint []string
}

func NewEngine(client client.APIClient, config Config) *Engine {
	return &Engine{
		config: config,
		client: client,
	}
}

func (e *Engine) Run(ctx context.Context, cc *ContainerConfig, writer io.Writer) (*types.ContainerState, error) {
	if err := e.create(ctx, cc); err != nil {
		return nil, err
	}

	if err := e.start(ctx, cc.Name); err != nil {
		return nil, err
	}

	if err := e.logs(ctx, cc.Name, writer); err != nil {
		return nil, err
	}

	return e.wait(ctx, cc.Name)
}

func (e *Engine) create(ctx context.Context, cc *ContainerConfig) error {
	var config *container.Config
	var hostConfig *container.HostConfig
	var networkingConfig *network.NetworkingConfig

	config = &container.Config{
		Image: cc.Image,
		// WorkingDir:   cc.WorkingDir,
		// Entrypoint:   cc.Entrypoint,
		AttachStdin:  false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
		OpenStdin:    false,
		StdinOnce:    false,
		ArgsEscaped:  false,
	}

	hostConfig = &container.HostConfig{
		Privileged: false,
	}

	networkingConfig = &network.NetworkingConfig{}

	_, err := e.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkingConfig,
		nil,
		cc.Name,
	)

	return err
}

func (e *Engine) start(ctx context.Context, id string) error {
	return e.client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (e *Engine) wait(ctx context.Context, id string) (*types.ContainerState, error) {
	wait, errc := e.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case <-wait:
	case <-errc:
	}

	info, err := e.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return info.State, nil
}

func (e *Engine) state(ctx context.Context, id string) (*types.ContainerState, error) {
	info, err := e.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return info.State, nil
}

func (e *Engine) logs(ctx context.Context, id string, writer io.Writer) error {
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
		Details:    false,
	}

	logs, err := e.client.ContainerLogs(ctx, id, logOptions)
	if err != nil {
		return err
	}

	go func() {
		stdcopy.StdCopy(writer, writer, logs)
		logs.Close()
	}()

	return nil
}
