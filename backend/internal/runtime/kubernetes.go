package runtime

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// KubernetesRuntime implements Runtime using the Kubernetes API.
type KubernetesRuntime struct {
	clientset       *kubernetes.Clientset
	workerNamespace string
}

// NewKubernetesRuntime creates a new Kubernetes-based runtime.
func NewKubernetesRuntime(cfg KubernetesConfig) (*KubernetesRuntime, error) {
	var config *rest.Config
	var err error

	if cfg.Kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	} else {
		// Try in-cluster first, fall back to default kubeconfig
		config, err = rest.InClusterConfig()
		if err != nil {
			config, err = clientcmd.BuildConfigFromFlags("", clientcmd.RecommendedHomeFile)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("build k8s config: %w", err)
	}

	cs, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("create k8s clientset: %w", err)
	}

	ns := cfg.WorkerNamespace
	if ns == "" {
		ns = cfg.Namespace
	}

	return &KubernetesRuntime{
		clientset:       cs,
		workerNamespace: ns,
	}, nil
}

func (k *KubernetesRuntime) Start(ctx context.Context, spec ContainerSpec) (ContainerID, error) {
	envVars := make([]corev1.EnvVar, 0, len(spec.Env))
	for key, val := range spec.Env {
		envVars = append(envVars, corev1.EnvVar{Name: key, Value: val})
	}

	ctr := corev1.Container{
		Name:      spec.Name,
		Image:     spec.Image,
		Env:       envVars,
		Command:   spec.Command,
		Args:      spec.Args,
		Resources: buildResourceRequirements(spec.Resources),
	}

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: k.workerNamespace,
			Labels:    spec.Labels,
		},
		Spec: corev1.PodSpec{
			RestartPolicy: corev1.RestartPolicyNever,
			Containers:    []corev1.Container{ctr},
		},
	}

	created, err := k.clientset.CoreV1().Pods(k.workerNamespace).Create(ctx, pod, metav1.CreateOptions{})
	if err != nil {
		return "", fmt.Errorf("create pod: %w", err)
	}

	return ContainerID(created.Name), nil
}

func (k *KubernetesRuntime) Stop(ctx context.Context, id ContainerID, timeout time.Duration) error {
	gracePeriod := int64(timeout.Seconds())
	err := k.clientset.CoreV1().Pods(k.workerNamespace).Delete(ctx, string(id), metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	})
	if err != nil {
		return fmt.Errorf("delete pod: %w", err)
	}
	return nil
}

func (k *KubernetesRuntime) Remove(ctx context.Context, id ContainerID) error {
	err := k.clientset.CoreV1().Pods(k.workerNamespace).Delete(ctx, string(id), metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("remove pod: %w", err)
	}
	return nil
}

func (k *KubernetesRuntime) Status(ctx context.Context, id ContainerID) (ContainerInfo, error) {
	pod, err := k.clientset.CoreV1().Pods(k.workerNamespace).Get(ctx, string(id), metav1.GetOptions{})
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("get pod: %w", err)
	}
	return podToContainerInfo(pod), nil
}

func (k *KubernetesRuntime) Logs(ctx context.Context, id ContainerID, opts LogOptions) (io.ReadCloser, error) {
	podLogOpts := &corev1.PodLogOptions{
		Follow:     opts.Follow,
		Timestamps: opts.Timestamps,
	}
	if opts.Tail > 0 {
		lines := int64(opts.Tail)
		podLogOpts.TailLines = &lines
	}
	if !opts.Since.IsZero() {
		sinceTime := metav1.NewTime(opts.Since)
		podLogOpts.SinceTime = &sinceTime
	}

	req := k.clientset.CoreV1().Pods(k.workerNamespace).GetLogs(string(id), podLogOpts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pod logs: %w", err)
	}
	return stream, nil
}

func (k *KubernetesRuntime) Wait(ctx context.Context, id ContainerID) (ContainerInfo, error) {
	watcher, err := k.clientset.CoreV1().Pods(k.workerNamespace).Watch(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("metadata.name=%s", string(id)),
	})
	if err != nil {
		return ContainerInfo{}, fmt.Errorf("watch pod: %w", err)
	}
	defer watcher.Stop()

	for event := range watcher.ResultChan() {
		if event.Type == watch.Error {
			return ContainerInfo{}, fmt.Errorf("watch error")
		}
		pod, ok := event.Object.(*corev1.Pod)
		if !ok {
			continue
		}
		switch pod.Status.Phase {
		case corev1.PodSucceeded, corev1.PodFailed:
			return podToContainerInfo(pod), nil
		}
	}

	return ContainerInfo{}, fmt.Errorf("watch closed unexpectedly")
}

func (k *KubernetesRuntime) List(ctx context.Context, labels map[string]string) ([]ContainerInfo, error) {
	parts := make([]string, 0, len(labels))
	for key, val := range labels {
		parts = append(parts, key+"="+val)
	}
	labelSelector := strings.Join(parts, ",")

	pods, err := k.clientset.CoreV1().Pods(k.workerNamespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("list pods: %w", err)
	}

	infos := make([]ContainerInfo, len(pods.Items))
	for i := range pods.Items {
		infos[i] = podToContainerInfo(&pods.Items[i])
	}
	return infos, nil
}

func (k *KubernetesRuntime) Healthy(ctx context.Context) error {
	_, err := k.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("k8s health check: %w", err)
	}
	return nil
}

func buildResourceRequirements(res ResourceRequirements) corev1.ResourceRequirements {
	reqs := corev1.ResourceRequirements{}
	if res.CPULimit != "" || res.MemoryLimit != "" {
		reqs.Limits = corev1.ResourceList{}
		if res.CPULimit != "" {
			reqs.Limits[corev1.ResourceCPU] = resource.MustParse(res.CPULimit)
		}
		if res.MemoryLimit != "" {
			reqs.Limits[corev1.ResourceMemory] = resource.MustParse(res.MemoryLimit)
		}
	}
	if res.CPURequest != "" || res.MemoryRequest != "" {
		reqs.Requests = corev1.ResourceList{}
		if res.CPURequest != "" {
			reqs.Requests[corev1.ResourceCPU] = resource.MustParse(res.CPURequest)
		}
		if res.MemoryRequest != "" {
			reqs.Requests[corev1.ResourceMemory] = resource.MustParse(res.MemoryRequest)
		}
	}
	return reqs
}

func podToContainerInfo(pod *corev1.Pod) ContainerInfo {
	ci := ContainerInfo{
		ID:   ContainerID(pod.Name),
		Name: pod.Name,
	}

	switch pod.Status.Phase {
	case corev1.PodPending:
		ci.Status = StatusPending
	case corev1.PodRunning:
		ci.Status = StatusRunning
	case corev1.PodSucceeded:
		ci.Status = StatusSucceeded
	case corev1.PodFailed:
		ci.Status = StatusFailed
	default:
		ci.Status = StatusUnknown
	}

	if pod.Status.StartTime != nil {
		ci.StartedAt = pod.Status.StartTime.Time
	}

	// Get exit code from first container status
	if len(pod.Status.ContainerStatuses) > 0 {
		cs := pod.Status.ContainerStatuses[0]
		if cs.State.Terminated != nil {
			ci.ExitCode = int(cs.State.Terminated.ExitCode)
			ci.FinishedAt = cs.State.Terminated.FinishedAt.Time
			ci.Message = cs.State.Terminated.Message
		}
	}

	return ci
}

