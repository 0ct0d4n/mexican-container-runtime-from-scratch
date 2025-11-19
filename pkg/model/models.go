package model

// UTS namespace: hostname y dominio
type UTSNamespace struct {
	Hostname string `json:"hostname"`
	Domain   string `json:"domain,omitempty"`
}

// PID namespace: aislamiento de procesos
type PIDNamespace struct {
	InitCommand []string `json:"init_command"` // proceso PID 1
}

// Network namespace: interfaces, rutas, bridges, etc.
type NetworkNamespace struct {
	Interfaces     []string          `json:"interfaces,omitempty"`
	Routes         map[string]string `json:"routes,omitempty"`
	EnableLoopback bool              `json:"enable_loopback"`
}

// Mount namespace: define el rootfs y sus mounts
type MountNamespace struct {
	Rootfs   string   `json:"rootfs"`
	Mounts   []string `json:"mounts,omitempty"` // ej: /proc, /sys, /dev
	ReadOnly bool     `json:"read_only"`
}

// IPC namespace: recursos de comunicación entre procesos
type IPCNamespace struct {
	SharedMemoryLimit int64 `json:"shm_limit"` // bytes
}

// User namespace: mapeo de UID/GID
type UserNamespace struct {
	UIDMap map[int]int `json:"uid_map"`
	GIDMap map[int]int `json:"gid_map"`
}

// Cgroup namespace: visibilidad y aislamiento de cgroups
type CgroupNamespace struct {
	Path      string  `json:"path"`
	MemoryMax uint64  `json:"memory_max_bytes"`
	CPUMax    float64 `json:"cpu_max_bytes"` // porcentaje o núcleos (opcional)
	CPUQuota  int64   `json:"cpu_quota"`     // µs (opcional)
	CPUPeriod int64   `json:"cpu_period"`    // µs (opcional)
	PidsMax   uint64  `json:"pids_max"`
}

// Time namespace: control del reloj
type TimeNamespace struct {
	OffsetSeconds int64 `json:"offset_seconds"`
}

type NamespaceConfig struct {
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
	Command   Commands         `json:"command,omitempty"`
	Namespace *NamespaceConfig `json:"namespaces,omitempty"`
}
type Commands struct {
	Command string
	Args    []string
}
