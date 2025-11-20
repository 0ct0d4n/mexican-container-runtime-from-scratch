package cgroups

import (
	"axolotl/libaxolotl/types"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	CgroupRoot = "/sys/fs/cgroup/axolotl/"
)

type Manager struct {
	path   string
	config *types.CgroupNamespace
}

func NewManager(cfg *types.ContainerSetupSettings) (*Manager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("configuration cannot be nil")
	}
	if cfg.Cgroup == nil {
		return nil, fmt.Errorf("cgroup configuration cannot be nil")
	}
	if cfg.Cgroup.Path == "" {
		return nil, fmt.Errorf("cgroup path cannot be empty")
	}
	if cfg.Org == "" {
		return nil, fmt.Errorf("org cannot be empty")
	}
	if cfg.ContainerName == "" {
		return nil, fmt.Errorf("container name cannot be empty")
	}

	cgroupPath := buildCgroupPath(cfg)
	log.Printf("[CGROUP] Creating cgroup at %s", cgroupPath)

	return &Manager{
		path:   cgroupPath,
		config: cfg.Cgroup,
	}, nil
}

func EnsureAxolotlRoot() error {
	if err := os.MkdirAll(CgroupRoot, 0755); err != nil {
		return err
	}

	procs, err := os.ReadFile(filepath.Join(CgroupRoot, "cgroup.procs"))
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(procs)) != "" {
		return fmt.Errorf("axolotl root cgroup has processes; can't enable controllers")
	}

	if err := os.WriteFile(
		filepath.Join(CgroupRoot, "cgroup.subtree_control"),
		[]byte("+memory +cpu +pids"),
		0644,
	); err != nil {
		return fmt.Errorf("enabling controllers on axolotl cgroup: %w", err)
	}
	return nil
}

func (m *Manager) CreateCgroup() error {
	info, err := os.Stat(m.path)

	if err == nil {
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", m.path)
		}
		return nil
	}

	if os.IsNotExist(err) {
		if err := os.MkdirAll(m.path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", m.path, err)
		}
		return nil
	}

	return fmt.Errorf("error checking %s: %w", m.path, err)
}

func (m *Manager) ApplyLimits() error {
	if err := m.applyMemoryLimit(); err != nil {
		return err
	}
	if err := m.applyCPULimit(); err != nil {
		return err
	}
	if err := m.applyPidsLimit(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) Path() string {
	return m.path
}

func (m *Manager) Destroy() error {
	if err := os.RemoveAll(m.path); err != nil {
		return fmt.Errorf("failed to remove cgroup %s: %w", m.path, err)
	}
	log.Printf("[CGROUP] Destroyed %s", m.path)
	return nil
}

func buildCgroupPath(cfg *types.ContainerSetupSettings) string {
	return filepath.Join(CgroupRoot, cfg.ID)
}

func currentSlice() string {
	data, err := os.ReadFile("/proc/self/cgroup")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}

		path := parts[2]

		if strings.Contains(path, ".slice") {
			return strings.TrimPrefix(path, "/")
		}
	}

	return ""
}

func RunningInsideSystemd() (bool, string) {
	content := currentSlice()

	if strings.Contains(content, "system.slice") {
		return true, content
	}

	if strings.Contains(content, "user.slice") {
		return true, content
	}

	return false, content
}
