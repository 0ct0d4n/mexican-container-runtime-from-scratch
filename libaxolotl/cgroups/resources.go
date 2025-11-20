package cgroups

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// applyMemoryLimit configures the memory.max cgroup controller
func (m *Manager) applyMemoryLimit() error {
	if m.config.MemoryMax <= 0 {
		log.Printf("[CGROUP] Skipping memory limit (no limit configured)")
		return nil
	}

	memoryMaxPath := filepath.Join(m.path, "memory.max")
	data := fmt.Sprintf("%d", m.config.MemoryMax)

	if err := os.WriteFile(memoryMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", memoryMaxPath, err)
	}

	memMB := float64(m.config.MemoryMax) / (1024 * 1024)
	log.Printf("[CGROUP] Memory limit applied: %.1f MB (%d bytes) at %s",
		memMB, m.config.MemoryMax, memoryMaxPath)

	return nil
}

// applyCPULimit configures the cpu.max cgroup controller
func (m *Manager) applyCPULimit() error {
	var (
		period int64 = 100000 // standard base period (100ms)
		quota  int64
	)

	switch {
	case m.config.CPUMax > 0:
		quota = int64(m.config.CPUMax * float64(period))

	case m.config.CPUPeriod > 0 && m.config.CPUQuota > 0:
		period = m.config.CPUPeriod
		quota = m.config.CPUQuota

	default:
		log.Printf("[CGROUP] Skipping CPU limit (no limit configured)")
		return nil
	}

	if quota <= 0 {
		log.Printf("[CGROUP] Warning: invalid CPU quota value %d (CPUMax=%.2f), skipping", quota, m.config.CPUMax)
		return nil
	}

	cpuMaxPath := filepath.Join(m.path, "cpu.max")
	data := fmt.Sprintf("%d %d", quota, period)

	if err := os.WriteFile(cpuMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", cpuMaxPath, err)
	}

	cpuPercent := (float64(quota) / float64(period)) * 100
	log.Printf("[CGROUP] CPU limit applied: %.2f%% (quota=%d period=%d) at %s",
		cpuPercent, quota, period, cpuMaxPath)

	return nil
}

// applyPidsLimit configures the pids.max cgroup controller
func (m *Manager) applyPidsLimit() error {
	if m.config.PidsMax <= 0 {
		log.Printf("[CGROUP] Skipping PIDs limit (no limit configured)")
		return nil
	}

	pidsMaxPath := filepath.Join(m.path, "pids.max")
	data := fmt.Sprintf("%d", m.config.PidsMax)

	if err := os.WriteFile(pidsMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", pidsMaxPath, err)
	}

	log.Printf("[CGROUP] PIDs limit applied: max %d processes at %s",
		m.config.PidsMax, pidsMaxPath)

	return nil
}

// AddProcess adds a process to this cgroup
func (m *Manager) AddProcess(pid int) error {
	procsPath := filepath.Join(m.path, "cgroup.procs")
	data := fmt.Sprintf("%d", pid)

	if err := os.WriteFile(procsPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to add process %d to cgroup: %w", pid, err)
	}

	log.Printf("[CGROUP] Process %d added to cgroup at %s", pid, m.path)
	return nil
}
