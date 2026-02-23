package runtime

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// DockerRuntime implements Runtime using the Docker SDK.
type DockerRuntime struct {
	client      *client.Client
	networkName string
}

// NewDockerRuntime creates a new Docker-based runtime.
func NewDockerRuntime(cfg DockerConfig) (*DockerRuntime, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}
	return &DockerRuntime{
		client:      cli,
		networkName: cfg.NetworkName,
	}, nil
}

func (d *DockerRuntime) Start(ctx context.Context, spec ContainerSpec) (ContainerID, error) {
	// Pull image if not present
	_, _, err := d.client.ImageInspectWithRaw(ctx, spec.Image)
	if err != nil {
		reader, pullErr := d.client.ImagePull(ctx, spec.Image, image.PullOptions{})
		if pullErr != nil {
			return "", fmt.Errorf("pull image %s: %w", spec.Image, pullErr)
		}
		defer reader.Close()
		// Drain the reader to complete the pull
		io.Copy(io.Discard, reader)
	}

	// Convert env map to slice
	envSlice := make([]string, 0, len(spec.Env))
	for k, v := range spec.Env {
		envSlice = append(envSlice, k+"="+v)
	}

	// Create container
	containerCfg := &container.Config{
		Image:  spec.Image,
		Env:    envSlice,
		Labels: spec.Labels,
	}
	if len(spec.Command) > 0 {
		containerCfg.Entrypoint = spec.Command
	}
	if len(spec.Args) > 0 {
		containerCfg.Cmd = spec.Args
	}

	hostCfg := &container.HostConfig{}
	if mem := parseMemoryBytes(spec.Resources.MemoryLimit); mem > 0 {
		hostCfg.Resources.Memory = mem
	}
	if cpu := parseCPUNano(spec.Resources.CPULimit); cpu > 0 {
		hostCfg.Resources.NanoCPUs = cpu
	}
	if len(spec.Capabilities) > 0 {
		hostCfg.CapAdd = spec.Capabilities
	}

	networkingCfg := &network.NetworkingConfig{}
	if d.networkName != "" {
		networkingCfg.EndpointsConfig = map[string]*network.EndpointSettings{
			d.networkName: {},
		}
	}

	resp, err := d.client.ContainerCreate(ctx, containerCfg, hostCfg, networkingCfg, nil, spec.Name)
	if err != nil {
		return "", fmt.Errorf("create container: %w", err)
	}

	// Start container
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		// Clean up on failure
		d.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{Force: true})
		return "", fmt.Errorf("start container: %w", err)
	}

	return ContainerID(resp.ID), nil
}

func (d *DockerRuntime) Stop(ctx context.Context, id ContainerID, timeout time.Duration) error {
	timeoutSec := int(timeout.Seconds())
	opts := container.StopOptions{Timeout: &timeoutSec}
	if err := d.client.ContainerStop(ctx, string(id), opts); err != nil {
		return fmt.Errorf("stop container: %w", err)
	}
	return nil
}

func (d *DockerRuntime) Remove(ctx context.Context, id ContainerID) error {
	if err := d.client.ContainerRemove(ctx, string(id), container.RemoveOptions{Force: true}); err != nil {
		return fmt.Errorf("remove container: %w", err)
	}
	return nil
}

func (d *DockerRuntime) Status(ctx context.Context, id ContainerID) (ContainerInfo, error) {
	info, err := d.client.ContainerInspect(ctx, string(id))
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("inspect container: %w", err)
	}
	return dockerInfoToContainerInfo(info), nil
}

func (d *DockerRuntime) Logs(ctx context.Context, id ContainerID, opts LogOptions) (io.ReadCloser, error) {
	logOpts := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
	}
	if opts.Tail > 0 {
		logOpts.Tail = fmt.Sprintf("%d", opts.Tail)
	}
	if !opts.Since.IsZero() {
		logOpts.Since = opts.Since.Format(time.RFC3339)
	}

	reader, err := d.client.ContainerLogs(ctx, string(id), logOpts)
	if err != nil {
		return nil, fmt.Errorf("get container logs: %w", err)
	}

	// Docker multiplexes stdout/stderr with 8-byte binary frame headers when
	// TTY is disabled (our default). Demux so callers receive clean UTF-8 text.
	pr, pw := io.Pipe()
	go func() {
		_, copyErr := stdcopy.StdCopy(pw, pw, reader)
		reader.Close()
		pw.CloseWithError(copyErr)
	}()
	return pr, nil
}

func (d *DockerRuntime) Wait(ctx context.Context, id ContainerID) (ContainerInfo, error) {
	statusCh, errCh := d.client.ContainerWait(ctx, string(id), container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			return ContainerInfo{}, fmt.Errorf("wait for container: %w", err)
		}
	case status := <-statusCh:
		info := ContainerInfo{
			ID:       id,
			ExitCode: int(status.StatusCode),
		}
		if status.StatusCode == 0 {
			info.Status = StatusSucceeded
		} else {
			info.Status = StatusFailed
		}
		if status.Error != nil {
			info.Message = status.Error.Message
		}
		return info, nil
	case <-ctx.Done():
		return ContainerInfo{}, ctx.Err()
	}
	return ContainerInfo{}, nil
}

func (d *DockerRuntime) List(ctx context.Context, labels map[string]string) ([]ContainerInfo, error) {
	filterArgs := filters.NewArgs()
	for k, v := range labels {
		filterArgs.Add("label", k+"="+v)
	}

	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		Filters: filterArgs,
		All:     true,
	})
	if err != nil {
		return nil, fmt.Errorf("list containers: %w", err)
	}

	infos := make([]ContainerInfo, len(containers))
	for i, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}
		infos[i] = ContainerInfo{
			ID:     ContainerID(c.ID),
			Name:   name,
			Status: dockerStateToStatus(c.State),
		}
	}
	return infos, nil
}

func (d *DockerRuntime) Healthy(ctx context.Context) error {
	_, err := d.client.Ping(ctx)
	if err != nil {
		return fmt.Errorf("docker ping: %w", err)
	}
	return nil
}

func dockerStateToStatus(state string) ContainerStatus {
	switch state {
	case "created":
		return StatusPending
	case "running":
		return StatusRunning
	case "exited":
		return StatusSucceeded // Caller should check exit code
	default:
		return StatusUnknown
	}
}

func dockerInfoToContainerInfo(info types.ContainerJSON) ContainerInfo {
	ci := ContainerInfo{
		ID:   ContainerID(info.ID),
		Name: strings.TrimPrefix(info.Name, "/"),
	}

	if info.State != nil {
		switch {
		case info.State.Running:
			ci.Status = StatusRunning
		case info.State.ExitCode == 0:
			ci.Status = StatusSucceeded
		case info.State.ExitCode != 0:
			ci.Status = StatusFailed
		default:
			ci.Status = StatusPending
		}
		ci.ExitCode = info.State.ExitCode

		if startedAt, err := time.Parse(time.RFC3339Nano, info.State.StartedAt); err == nil {
			ci.StartedAt = startedAt
		}
		if finishedAt, err := time.Parse(time.RFC3339Nano, info.State.FinishedAt); err == nil {
			ci.FinishedAt = finishedAt
		}
	}

	return ci
}

// parseMemoryBytes converts a Kubernetes-style memory string (e.g. "256Mi",
// "1Gi", "512000") to bytes. Returns 0 on empty or unparseable input.
func parseMemoryBytes(s string) int64 {
	if s == "" {
		return 0
	}
	multipliers := map[string]int64{
		"Ki": 1024,
		"Mi": 1024 * 1024,
		"Gi": 1024 * 1024 * 1024,
		"Ti": 1024 * 1024 * 1024 * 1024,
		"K":  1000,
		"M":  1000 * 1000,
		"G":  1000 * 1000 * 1000,
	}
	for suffix, mult := range multipliers {
		if strings.HasSuffix(s, suffix) {
			numStr := strings.TrimSuffix(s, suffix)
			n, err := strconv.ParseInt(numStr, 10, 64)
			if err != nil {
				return 0
			}
			return n * mult
		}
	}
	// Plain number = bytes.
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

// parseCPUNano converts a Kubernetes-style CPU string (e.g. "500m", "2") to
// Docker NanoCPUs (1 CPU = 1e9 NanoCPUs). Returns 0 on empty or unparseable input.
func parseCPUNano(s string) int64 {
	if s == "" {
		return 0
	}
	if strings.HasSuffix(s, "m") {
		numStr := strings.TrimSuffix(s, "m")
		n, err := strconv.ParseInt(numStr, 10, 64)
		if err != nil {
			return 0
		}
		return n * 1_000_000 // 1 milli-CPU = 1e6 NanoCPUs
	}
	n, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return int64(n * 1e9)
}
