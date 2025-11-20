// Package cgroups provides cgroup v2 management for container resource isolation
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
	// CgroupRoot is the cgroup v2 unified hierarchy mount point
	CgroupRoot = "/sys/fs/cgroup/axolotl/"
)

// Manager handles cgroup lifecycle and resource limits
type Manager struct {
	path   string
	config *types.CgroupNamespace
}

// NewManager creates a new cgroup manager for the given container configuration
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

	if err := ensureDir(cgroupPath); err != nil {
		return nil, fmt.Errorf("failed to create cgroup directory: %w", err)
	}

	log.Printf("[CGROUP] Created successfully at %s", cgroupPath)

	return &Manager{
		path:   cgroupPath,
		config: cfg.Cgroup,
	}, nil
}

// ApplyLimits applies all configured resource limits (memory, CPU, PIDs)
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

// Path returns the cgroup filesystem path
func (m *Manager) Path() string {
	return m.path
}

// Destroy removes the cgroup directory
func (m *Manager) Destroy() error {
	if err := os.RemoveAll(m.path); err != nil {
		return fmt.Errorf("failed to remove cgroup %s: %w", m.path, err)
	}
	log.Printf("[CGROUP] Destroyed %s", m.path)
	return nil
}

// buildCgroupPath constructs the cgroup path from container settings
func buildCgroupPath(cfg *types.ContainerSetupSettings) string {
	// Use underscore separator to create a flat cgroup hierarchy
	// Example: /sys/fs/cgroup/org_containername_path
	//name := cfg.Org + "_" + cfg.ContainerName + "_" + cfg.Cgroup.Path
	return filepath.Join(CgroupRoot, cfg.ID)
}

// ensureDir creates the directory if it doesn't exist, or verifies it's a directory
func ensureDir(path string) error {
	info, err := os.Stat(path)

	if err == nil {
		// Path exists: verify it's a directory
		if !info.IsDir() {
			return fmt.Errorf("%s exists but is not a directory", path)
		}
		return nil
	}

	// If doesn't exist, create it
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", path, err)
		}
		return nil
	}

	// Unexpected error
	return fmt.Errorf("error checking %s: %w", path, err)
}

// currentSlice returns the current systemd slice this process belongs to
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

		// This is the cgroup path
		path := parts[2] // example: "/system.slice/axod.service"

		// Identify valid systemd paths
		if strings.Contains(path, ".slice") {
			// Remove leading "/"
			return strings.TrimPrefix(path, "/")
		}
	}

	return ""
}

// RunningInsideSystemd checks if the process is running inside systemd
func RunningInsideSystemd() (bool, string) {
	content := currentSlice()

	// systemd always puts services in "system.slice/<service>.service"
	if strings.Contains(content, "system.slice") {
		return true, content
	}

	// user services: "user.slice/user-1000.slice"
	if strings.Contains(content, "user.slice") {
		return true, content
	}

	// if it doesn't belong to any "usual" slice, it's normal execution
	return false, content
}
