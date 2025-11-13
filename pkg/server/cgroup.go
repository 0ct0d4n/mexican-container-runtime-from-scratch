package server

import (
	"axolotl/pkg/model"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const CgroupSysPath = "/sys/fs/cgroup/"
const CgroupSysPathNS = CgroupSysPath + "axolotl"

type TmpCgroup struct {
	path   string
	cgroup *model.CgroupNamespace
}

func createCgroup(cfg *model.NamespaceConfig) (*TmpCgroup, error) {
	// Validación de entrada
	if cfg == nil {
		return nil, fmt.Errorf("❌ [CGROUP] Configuración nula")
	}
	if cfg.Cgroup.Path == "" {
		return nil, fmt.Errorf("❌ [CGROUP] No se especificó ruta para el cgroup")
	}
	if cfg.Org == "" {
		return nil, fmt.Errorf("❌ [CGROUP] No se especificó Org para el cgroup")
	}
	if cfg.ContainerName == "" {
		return nil, fmt.Errorf("❌ [CGROUP] No se especificó ContainerName para el cgroup")
	}

	tmpPath := BuildCgroupPath(cfg)
	log.Printf("[CGROUP] Creating cgroup at %s", tmpPath)
	if err := EnsureDir(tmpPath); err != nil {
		return nil, err
	}
	log.Printf("[CGROUP] Created successfully at %s", tmpPath)
	// Devolver instancia temporal
	return &TmpCgroup{
		path:   tmpPath,
		cgroup: cfg.Cgroup,
	}, nil
}
func (t *TmpCgroup) LimitResources() error {
	err := t.writeMemoryMax()
	if err != nil {
		return err
	}
	err = t.writeCPUMax()
	if err != nil {
		return err
	}
	err = t.writePidsMax()
	if err != nil {
		return err
	}
	return nil
}

func (t *TmpCgroup) writeMemoryMax() error {

	if t.cgroup.MemoryMax <= 0 {
		fmt.Println("🦎 [MEM] Sin límite configurado, omitiendo memory.max")
		return nil
	}
	memoryMaxPath := filepath.Join(t.path, "memory.max")
	data := fmt.Sprintf("%d", t.cgroup.MemoryMax)

	if err := os.WriteFile(memoryMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("❌ [MEM] Error escribiendo %s: %w", memoryMaxPath, err)
	}

	memMB := float64(t.cgroup.MemoryMax) / (1024 * 1024)
	fmt.Printf("✅ [MEM] memory.max configurado en %s → %.1f MB (%d bytes)\n",
		memoryMaxPath, memMB, t.cgroup.MemoryMax)

	return nil
}

func (t *TmpCgroup) writeCPUMax() error {
	var (
		period int64 = 100000 // período base estándar (100ms)
		quota  int64
	)

	switch {
	case t.cgroup.CPUMax > 0:
		quota = int64(t.cgroup.CPUMax * float64(period))

	case t.cgroup.CPUPeriod > 0 && t.cgroup.CPUQuota > 0:
		period = t.cgroup.CPUPeriod
		quota = t.cgroup.CPUQuota

	default:
		fmt.Println("🦎 [CPU] Sin límite configurado, omitiendo cpu.max")
		return nil
	}

	if quota <= 0 {
		fmt.Printf("⚠️ [CPU] Valor inválido de quota: %d (CPUMax=%.2f)\n", quota, t.cgroup.CPUMax)
		return nil
	}

	cpuMaxPath := filepath.Join(t.path, "cpu.max")

	data := fmt.Sprintf("%d %d", quota, period)

	if err := os.WriteFile(cpuMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("❌ [CPU] Error escribiendo %s: %w", cpuMaxPath, err)
	}

	fmt.Printf("✅ [CPU] cpu.max configurado en %s → quota=%d period=%d (%.2f%%)\n",
		cpuMaxPath, quota, period, (float64(quota)/float64(period))*100)

	return nil
}

func (t *TmpCgroup) writePidsMax() error {
	pidsMaxPath := filepath.Join(t.path, "pids.max")

	if t.cgroup.PidsMax <= 0 {
		fmt.Println("🦎 [PIDS] Sin límite configurado, omitiendo pids.max")
		return nil
	}

	data := fmt.Sprintf("%d", t.cgroup.PidsMax)
	if err := os.WriteFile(pidsMaxPath, []byte(data), 0644); err != nil {
		return fmt.Errorf("❌ [PIDS] Error escribiendo %s: %w", pidsMaxPath, err)
	}

	fmt.Printf("✅ [PIDS] pids.max configurado en %s → máximo %d procesos\n",
		pidsMaxPath, t.cgroup.PidsMax)

	return nil
}
