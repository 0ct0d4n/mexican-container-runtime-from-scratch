package types

// UTSNamespace configures hostname and domain isolation
type UTSNamespace struct {
	Hostname string `json:"hostname"`
	Domain   string `json:"domain,omitempty"`
}

// PIDNamespace configures process isolation
type PIDNamespace struct {
	InitCommand []string `json:"init_command"` // PID 1 process
}

// NetworkNamespace configures network isolation (interfaces, routes, bridges)
type NetworkNamespace struct {
	Interfaces     []string          `json:"interfaces,omitempty"`
	Routes         map[string]string `json:"routes,omitempty"`
	EnableLoopback bool              `json:"enable_loopback"`
}

// MountNamespace defines the rootfs and mount points
type MountNamespace struct {
	Rootfs   string   `json:"rootfs"`
	Mounts   []string `json:"mounts,omitempty"` // e.g., /proc, /sys, /dev
	ReadOnly bool     `json:"read_only"`
}

// IPCNamespace configures inter-process communication resources
type IPCNamespace struct {
	SharedMemoryLimit int64 `json:"shm_limit"` // bytes
}

// UserNamespace configures UID/GID mapping
type UserNamespace struct {
	UIDMap map[int]int `json:"uid_map"`
	GIDMap map[int]int `json:"gid_map"`
}

// CgroupNamespace configures cgroup visibility and isolation
type CgroupNamespace struct {
	Path      string  `json:"path"`
	MemoryMax uint64  `json:"memory_max_bytes"`
	CPUMax    float64 `json:"cpu_max_bytes"` // percentage or cores (optional)
	CPUQuota  int64   `json:"cpu_quota"`     // microseconds (optional)
	CPUPeriod int64   `json:"cpu_period"`    // microseconds (optional)
	PidsMax   uint64  `json:"pids_max"`
}

// TimeNamespace controls clock isolation
type TimeNamespace struct {
	OffsetSeconds int64 `json:"offset_seconds"`
}

type ContainerSetupSettings struct {
	ID            string            `json:"id,omitempty"`
	Org           string            `json:"org,omitempty"`
	ContainerName string            `json:"container_name"`
	ImageName     string            `json:"image_name,omitempty"`
	UTS           *UTSNamespace     `json:"uts,omitempty"`
	PID           *PIDNamespace     `json:"pid,omitempty"`
	Network       *NetworkNamespace `json:"network,omitempty"`
	Mount         *MountNamespace   `json:"mount,omitempty"`
	IPC           *IPCNamespace     `json:"ipc,omitempty"`
	User          *UserNamespace    `json:"user,omitempty"`
	Cgroup        *CgroupNamespace  `json:"cgroup,omitempty"`
	Time          *TimeNamespace    `json:"time,omitempty"`
}

type RunRequest struct {
	Command   Commands                `json:"command,omitempty"`
	Namespace *ContainerSetupSettings `json:"namespaces,omitempty"`
}
type Commands struct {
	Command string
	Args    []string
}
