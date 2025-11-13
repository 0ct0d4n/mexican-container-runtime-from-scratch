package server

import (
	"axolotl/pkg/model"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func RunningInsideSystemd() (bool, string) {
	content := CurrentSlice()

	// systemd siempre mete servicios en "system.slice/<servicio>.service"
	if strings.Contains(content, "system.slice") {
		return true, content
	}

	// servicios de usuario: "user.slice/user-1000.slice"
	if strings.Contains(content, "user.slice") {
		return true, content
	}

	// si no pertenece a ningún slice "usual", es ejecución normal
	return false, content
}

func EnsureDir(path string) error {
	info, err := os.Stat(path)

	if err == nil {
		// El path existe: verificar que realmente sea un directorio
		if !info.IsDir() {
			return fmt.Errorf("❌ %s existe pero no es un directorio", path)
		}
		return nil
	}

	// Si no existe, lo creamos
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("❌ No se pudo crear directorio %s: %w", path, err)
		}
		return nil
	}

	// Otro error inesperado
	return fmt.Errorf("❌ Error verificando %s: %w", path, err)
}

func BuildCgroupPath(cfg *model.NamespaceConfig) string {
	return filepath.Join(CgroupSysPathNS, cfg.Org, cfg.ContainerName, cfg.Cgroup.Path)
}

func CurrentSlice() string {
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

		// este es el path del cgroup
		path := parts[2] // ejemplo: "/system.slice/axod.service"

		// identificamos rutas válidas de systemd
		if strings.Contains(path, ".slice") {
			// removemos el "/" inicial
			return strings.TrimPrefix(path, "/")
		}
	}

	return ""
}
