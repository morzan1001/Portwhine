package runtime

import (
	"context"
	"fmt"
	"io"
	"time"
)

// ContainerID is an opaque identifier for a running container/pod.
type ContainerID string

// ContainerStatus represents the lifecycle state of a container.
type ContainerStatus string

const (
	StatusPending   ContainerStatus = "pending"
	StatusRunning   ContainerStatus = "running"
	StatusSucceeded ContainerStatus = "succeeded"
	StatusFailed    ContainerStatus = "failed"
	StatusUnknown   ContainerStatus = "unknown"
)

// ContainerSpec defines what container to run.
type ContainerSpec struct {
	// Image is the container image reference (e.g., "portwhine/dns-resolver:latest").
	Image string
	// Name is a human-readable name for the container/pod.
	Name string
	// Command overrides the entrypoint.
	Command []string
	// Args passed to the entrypoint or command.
	Args []string
	// Env is a map of environment variables.
	Env map[string]string
	// Labels are metadata attached to the container/pod.
	Labels map[string]string
	// Resources define CPU and memory constraints.
	Resources ResourceRequirements
	// Network configuration for operator connectivity.
	Network NetworkConfig
	// Capabilities are Linux capabilities to add (e.g., "NET_ADMIN", "SYS_ADMIN").
	Capabilities []string
}

// ResourceRequirements specifies compute resource constraints.
type ResourceRequirements struct {
	CPULimit      string // e.g., "500m", "2"
	MemoryLimit   string // e.g., "256Mi", "1Gi"
	CPURequest    string
	MemoryRequest string
}

// NetworkConfig specifies how the container connects to the operator.
type NetworkConfig struct {
	// OperatorAddress is the address the worker uses to reach the operator gRPC server.
	OperatorAddress string
	// NetworkName is the Docker network to attach to (Docker-only, ignored in K8s).
	NetworkName string
}

// ContainerInfo provides the current state of a container.
type ContainerInfo struct {
	ID         ContainerID
	Name       string
	Status     ContainerStatus
	ExitCode   int
	StartedAt  time.Time
	FinishedAt time.Time
	Message    string
}

// LogOptions controls log streaming behavior.
type LogOptions struct {
	Follow     bool
	Since      time.Time
	Tail       int
	Timestamps bool
}

// Runtime is the central abstraction over container orchestration backends.
type Runtime interface {
	// Start creates and starts a container/pod from the given spec.
	Start(ctx context.Context, spec ContainerSpec) (ContainerID, error)

	// Stop gracefully stops a running container/pod.
	Stop(ctx context.Context, id ContainerID, timeout time.Duration) error

	// Remove removes a stopped container/pod and its resources.
	Remove(ctx context.Context, id ContainerID) error

	// Status returns the current status of a container/pod.
	Status(ctx context.Context, id ContainerID) (ContainerInfo, error)

	// Logs returns a reader for streaming container logs.
	Logs(ctx context.Context, id ContainerID, opts LogOptions) (io.ReadCloser, error)

	// Wait blocks until the container/pod exits.
	Wait(ctx context.Context, id ContainerID) (ContainerInfo, error)

	// List returns all containers/pods managed by Portwhine (filtered by labels).
	List(ctx context.Context, labels map[string]string) ([]ContainerInfo, error)

	// Healthy checks connectivity to the underlying runtime.
	Healthy(ctx context.Context) error
}

// RuntimeType identifies the container runtime backend.
type RuntimeType string

const (
	RuntimeDocker     RuntimeType = "docker"
	RuntimeKubernetes RuntimeType = "kubernetes"
)

// Config holds runtime-specific configuration.
type Config struct {
	Type       RuntimeType
	Docker     DockerConfig
	Kubernetes KubernetesConfig
}

// DockerConfig holds Docker-specific settings.
type DockerConfig struct {
	NetworkName string
	Host        string          // Remote Docker daemon address (empty = local socket)
	TLS         DockerTLSConfig // TLS settings for remote daemon connection
}

// DockerTLSConfig holds TLS certificate paths for connecting to a remote Docker daemon.
type DockerTLSConfig struct {
	CACert string // Path to CA certificate
	Cert   string // Path to client certificate
	Key    string // Path to client key
}

// RemoteAddressResolver is optionally implemented by runtimes that need
// IP-based container addressing instead of DNS-based resolution.
type RemoteAddressResolver interface {
	ContainerAddress(ctx context.Context, id ContainerID) (string, error)
	IsRemote() bool
}

// KubernetesConfig holds Kubernetes-specific settings.
type KubernetesConfig struct {
	Namespace       string
	WorkerNamespace string
	Kubeconfig      string
}

// New creates the appropriate Runtime based on configuration.
func New(cfg Config) (Runtime, error) {
	switch cfg.Type {
	case RuntimeDocker:
		return NewDockerRuntime(cfg.Docker)
	case RuntimeKubernetes:
		return NewKubernetesRuntime(cfg.Kubernetes)
	default:
		return nil, fmt.Errorf("unsupported runtime type: %s", cfg.Type)
	}
}
