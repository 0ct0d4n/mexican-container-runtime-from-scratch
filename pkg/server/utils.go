package server

import (
	"axolotl/pkg/model"
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

func BuildCgroupPath(cfg *model.CgroupNamespace) string {
	_, contentSlice := RunningInsideSystemd()
	return filepath.Join(CgroupSysPath, contentSlice, cfg.Path)
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
