//go:build linux
// +build linux

package rootfs

import (
	"axolotl/internal/util"
	"axolotl/libaxolotl/types"
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"strings"
)

type Container struct {
	ID           string
	InstallPath  string
	DownloadPath string
	Config       *Config
	Distro       DistroType
	RootfsPath   string
	Commands     types.Commands
}

func Install(basePath string, cfg *types.ContainerSetupSettings, commands types.Commands) (*Container, error) {
	log.Printf("[ROOTFS] Installing image: %s", cfg.ImageName)

	distroType := DistroType(strings.ToUpper(cfg.ImageName))
	distroCfg, err := DetectDistroConfig(distroType)
	if err != nil {
		return nil, fmt.Errorf("failed to detect distro config: %w", err)
	}

	installPath := filepath.Join(basePath, cfg.ImageName)
	downloadPath := filepath.Join(installPath, FilenameFromURL(distroCfg.URL))
	rootfsPath := filepath.Join(installPath, "rootfs")

	if err := Download(distroCfg.URL, downloadPath); err != nil {
		return nil, fmt.Errorf("failed to download rootfs: %w", err)
	}

	if err := util.UntarRootFS(downloadPath, rootfsPath); err != nil {
		return nil, fmt.Errorf("failed to extract rootfs: %w", err)
	}

	container := &Container{
		ID:           cfg.ID,
		InstallPath:  installPath,
		DownloadPath: downloadPath,
		Config:       &distroCfg,
		Distro:       distroType,
		RootfsPath:   rootfsPath,
		Commands:     commands,
	}

	log.Printf("[ROOTFS] Image installed at %s", rootfsPath)
	return container, nil
}

func (c *Container) Cleanup() {
	log.Printf("[ROOTFS] Cleaning up mounts for %s", c.RootfsPath)
	mounts := GetDefaultMounts(c.RootfsPath)
	UnmountAll(mounts)
}

func generateVethNames(containerID string) (string, string) {
	shortID := containerID
	if len(shortID) > 11 {
		shortID = shortID[:11]
	}
	return "veth-" + shortID, "vethc-" + shortID
}

func (c *Container) ValidateRootfs() error {
	if c.RootfsPath == "" {
		return errors.New("rootfs path is empty")
	}

	criticalDirs := []string{"/bin", "/etc"}
	for _, dir := range criticalDirs {
		fullPath := filepath.Join(c.RootfsPath, dir)
		if err := util.FileExists(fullPath); err != nil {
			return fmt.Errorf("critical directory missing: %s", dir)
		}
	}

	return nil
}
