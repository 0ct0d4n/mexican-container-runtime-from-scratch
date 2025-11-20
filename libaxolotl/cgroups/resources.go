package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
)

// applyMemoryLimit configures the memory.max cgroup controller
func (m *Manager) applyMemoryLimit() error {
	if m.config.MemoryMax <= 0 {
		fmt.Println("🦎 [MEM] No limit configured, skipping memory.max")
		return nil
	}

	memoryMaxPath := filepath.Join(m.path, "memory.max")
	data := fmt.Sprintf("%d", m.config.MemoryMax)

	if err := os.WriteFile(memoryMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", memoryMaxPath, err)
	}

	memMB := float64(m.config.MemoryMax) / (1024 * 1024)
	fmt.Printf("✅ [MEM] memory.max configured at %s → %.1f MB (%d bytes)\n",
		memoryMaxPath, memMB, m.config.MemoryMax)

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
		fmt.Println("🦎 [CPU] No limit configured, skipping cpu.max")
		return nil
	}

	if quota <= 0 {
		fmt.Printf("⚠️ [CPU] Invalid quota value: %d (CPUMax=%.2f)\n", quota, m.config.CPUMax)
		return nil
	}

	cpuMaxPath := filepath.Join(m.path, "cpu.max")
	data := fmt.Sprintf("%d %d", quota, period)

	if err := os.WriteFile(cpuMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", cpuMaxPath, err)
	}

	fmt.Printf("✅ [CPU] cpu.max configured at %s → quota=%d period=%d (%.2f%%)\n",
		cpuMaxPath, quota, period, (float64(quota)/float64(period))*100)

	return nil
}

// applyPidsLimit configures the pids.max cgroup controller
func (m *Manager) applyPidsLimit() error {
	if m.config.PidsMax <= 0 {
		fmt.Println("🦎 [PIDS] No limit configured, skipping pids.max")
		return nil
	}

	pidsMaxPath := filepath.Join(m.path, "pids.max")
	data := fmt.Sprintf("%d", m.config.PidsMax)

	if err := os.WriteFile(pidsMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", pidsMaxPath, err)
	}

	fmt.Printf("✅ [PIDS] pids.max configured at %s → maximum %d processes\n",
		pidsMaxPath, m.config.PidsMax)

	return nil
}

// AddProcess adds a process to this cgroup
func (m *Manager) AddProcess(pid int) error {
	procsPath := filepath.Join(m.path, "cgroup.procs")
	data := fmt.Sprintf("%d", pid)

	if err := os.WriteFile(procsPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to add process %d to cgroup: %w", pid, err)
	}

	fmt.Printf("✅ [CGROUP] Process %d added to cgroup %s\n", pid, m.path)
	return nil
}
