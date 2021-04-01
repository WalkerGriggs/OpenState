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

type Docker struct {
	config Config

	client client.APIClient
}

type ContainerConfig struct {
	Name       string
	Image      string
	WorkingDir string
	Entrypoint []string
}

func NewDocker(client client.APIClient, config Config) *Docker {
	return &Docker{
		config: config,
		client: client,
	}
}

func (d *Docker) Run(ctx context.Context, cc *ContainerConfig, writer io.Writer) (*types.ContainerState, error) {
	if err := d.create(ctx, cc); err != nil {
		return nil, err
	}

	if err := d.start(ctx, cc.Name); err != nil {
		return nil, err
	}

	if err := d.logs(ctx, cc.Name, writer); err != nil {
		return nil, err
	}

	return d.wait(ctx, cc.Name)
}

func (d *Docker) create(ctx context.Context, cc *ContainerConfig) error {
	var config *container.Config
	var hostConfig *container.HostConfig
	var networkingConfig *network.NetworkingConfig

	config = &container.Config{
		Image:       cc.Image,
		WorkingDir:  cc.WorkingDir,
		Entrypoint:  cc.Entrypoint,
		AttachStdin: false,
		AttachStdout: true,
		AttachStderr: true,
		Tty:         false,
		OpenStdin:   false,
		StdinOnce:   false,
		ArgsEscaped: false,
	}

	hostConfig = &container.HostConfig{
		Privileged: false,
	}

	networkingConfig = &network.NetworkingConfig{}

	_, err := d.client.ContainerCreate(
		ctx,
		config,
		hostConfig,
		networkingConfig,
		nil,
		cc.Name,
	)

	return err
}

func (d *Docker) start(ctx context.Context, id string) error {
	return d.client.ContainerStart(ctx, id, types.ContainerStartOptions{})
}

func (d *Docker) wait(ctx context.Context, id string) (*types.ContainerState, error) {
	wait, errc := d.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	select {
	case <-wait:
	case <-errc:
	}

	info, err := d.client.ContainerInspect(ctx, id)
	if err != nil {
		return nil, err
	}

	return info.State, nil
}

func (d *Docker) logs(ctx context.Context, id string, writer io.Writer) error {
	logOptions := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Timestamps: true,
		Follow:     true,
		Details:    false,
	}

	logs, err := d.client.ContainerLogs(ctx, id, logOptions)
	if err != nil {
		return err
		
	}

	go func() {
		stdcopy.StdCopy(writer, writer, logs)
		logs.Close()
	}()

	return nil
}
