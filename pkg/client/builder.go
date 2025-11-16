package client

import (
	"axolotl/pkg/model"
	"fmt"
	"github.com/docker/go-units"
)

// RequestBuilder builds a RunRequest with validation.
type RequestBuilder struct {
	command       model.Commands
	containerName string
	image         string
	org           string
	cgroup        *model.CgroupNamespace
	uts           *model.UTSNamespace
	pid           *model.PIDNamespace
	network       *model.NetworkNamespace
	mount         *model.MountNamespace
	ipc           *model.IPCNamespace
	user          *model.UserNamespace
	time          *model.TimeNamespace
}

// NewRequestBuilder creates a new RequestBuilder.
func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{}
}

// WithCommand sets the command to execute.
func (b *RequestBuilder) WithCommand(cmd model.Commands) *RequestBuilder {
	b.command = cmd
	return b
}
func (b *RequestBuilder) WithImage(imageName string) *RequestBuilder {
	b.image = imageName
	return b
}

// WithContainerName sets the container name.
func (b *RequestBuilder) WithContainerName(name string) *RequestBuilder {
	b.containerName = name
	return b
}

// WithContainerName sets the container name.
func (b *RequestBuilder) WithOrg(name string) *RequestBuilder {
	b.org = name
	return b
}

// WithCgroup configures cgroup namespace limits.
func (b *RequestBuilder) WithCgroup(path string, memoryMB int64, cpuPercent float64, maxPids uint64) *RequestBuilder {
	b.cgroup = &model.CgroupNamespace{
		Path:      path,
		MemoryMax: uint64(memoryMB * units.MB),
		CPUMax:    cpuPercent,
		PidsMax:   maxPids,
	}
	return b
}

// WithCgroupRaw sets the cgroup configuration directly.
func (b *RequestBuilder) WithCgroupRaw(cgroup *model.CgroupNamespace) *RequestBuilder {
	b.cgroup = cgroup
	return b
}

// WithUTS configures UTS namespace (hostname and domain).
func (b *RequestBuilder) WithUTS(hostname, domain string) *RequestBuilder {
	b.uts = &model.UTSNamespace{
		Hostname: hostname,
		Domain:   domain,
	}
	return b
}

// WithPID configures PID namespace.
func (b *RequestBuilder) WithPID(initCommand []string) *RequestBuilder {
	b.pid = &model.PIDNamespace{
		InitCommand: initCommand,
	}
	return b
}

// WithNetwork configures network namespace.
func (b *RequestBuilder) WithNetwork(enableLoopback bool, interfaces []string, routes map[string]string) *RequestBuilder {
	b.network = &model.NetworkNamespace{
		EnableLoopback: enableLoopback,
		Interfaces:     interfaces,
		Routes:         routes,
	}
	return b
}

// WithMount configures mount namespace.
func (b *RequestBuilder) WithMount(rootfs string, mounts []string, readOnly bool) *RequestBuilder {
	b.mount = &model.MountNamespace{
		Rootfs:   rootfs,
		Mounts:   mounts,
		ReadOnly: readOnly,
	}
	return b
}

// WithIPC configures IPC namespace.
func (b *RequestBuilder) WithIPC(shmLimit int64) *RequestBuilder {
	b.ipc = &model.IPCNamespace{
		SharedMemoryLimit: shmLimit,
	}
	return b
}

// WithUser configures user namespace.
func (b *RequestBuilder) WithUser(uidMap, gidMap map[int]int) *RequestBuilder {
	b.user = &model.UserNamespace{
		UIDMap: uidMap,
		GIDMap: gidMap,
	}
	return b
}

// WithTime configures time namespace.
func (b *RequestBuilder) WithTime(offsetSeconds int64) *RequestBuilder {
	b.time = &model.TimeNamespace{
		OffsetSeconds: offsetSeconds,
	}
	return b
}

// Build creates a RunRequest with validation.
func (b *RequestBuilder) Build() (*model.RunRequest, error) {
	// Build namespace config
	namespace := &model.NamespaceConfig{
		ContainerName: b.containerName,
		Org:           b.org,
		ImageName:     b.image,
		Cgroup:        b.cgroup,
		UTS:           b.uts,
		PID:           b.pid,
		Network:       b.network,
		Mount:         b.mount,
		IPC:           b.ipc,
		User:          b.user,
		Time:          b.time,
	}

	// Validate
	if err := b.validate(namespace); err != nil {
		return nil, err
	}

	return &model.RunRequest{
		Command:   b.command,
		Namespace: namespace,
	}, nil
}

// validate performs validation on the request.
func (b *RequestBuilder) validate(ns *model.NamespaceConfig) error {
	// At least one namespace configuration should be provided
	hasNamespace := ns.Cgroup != nil || ns.UTS != nil || ns.PID != nil ||
		ns.Network != nil || ns.Mount != nil || ns.IPC != nil ||
		ns.User != nil || ns.Time != nil

	if !hasNamespace {
		return fmt.Errorf("at least one namespace configuration must be provided")
	}

	// Validate cgroup if provided
	if ns.Cgroup != nil {
		if ns.Cgroup.Path == "" {
			return fmt.Errorf("cgroup path cannot be empty")
		}
		if ns.Cgroup.MemoryMax > 0 && ns.Cgroup.MemoryMax < 1*units.MB {
			return fmt.Errorf("memory limit too low (minimum 1MB)")
		}
		if ns.Cgroup.CPUMax < 0 || ns.Cgroup.CPUMax > 100 {
			return fmt.Errorf("CPU limit must be between 0 and 100")
		}
	}

	// Validate UTS if provided
	if ns.UTS != nil && ns.UTS.Hostname == "" {
		return fmt.Errorf("UTS hostname cannot be empty")
	}

	// Validate mount if provided
	if ns.Mount != nil && ns.Mount.Rootfs == "" {
		return fmt.Errorf("mount rootfs cannot be empty")
	}

	return nil
}

// BuildOrPanic creates a RunRequest or panics if validation fails.
// Useful for testing or known-good configurations.
func (b *RequestBuilder) BuildOrPanic() *model.RunRequest {
	req, err := b.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build request: %v", err))
	}
	return req
}
