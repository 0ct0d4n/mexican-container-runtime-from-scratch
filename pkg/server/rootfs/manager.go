//go:build linux
// +build linux

package rootfs

import (
	"axolotl/pkg/model"
	"axolotl/pkg/util"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"path/filepath"
	"syscall"
)
import "axolotl/pkg/util/rootfs"

type MountPoint struct {
	Src    string
	Dst    string
	Fstype string
	Flags  uintptr
	Data   string
}

// /var/axolotl/images/
type RootFSInstallationConfig struct {
	InstallationPath    string
	DownloadName        string
	Config              *rootfs.RootFSConfig
	Distro              rootfs.DistroType
	CanonicalRootfsPath string
}

func InstallImage(diskPath string, config *model.NamespaceConfig) (*RootFSInstallationConfig, error) {
	var distroCfg rootfs.RootFSConfig
	var err error

	distroType := rootfs.DistroType(config.ImageName)
	if distroCfg, err = rootfs.DetectHostDistroConfig(distroType); err != nil {
		log.Println("error detecting host distro config:", err)
		return nil, errors.New("error detecting host distro config")
	}
	// bajar
	installationPath := filepath.Join(diskPath, config.ImageName)
	downloadName := filepath.Join(installationPath, FilenameFromURL(distroCfg.URL))
	canonicalRootfsPath := filepath.Join(installationPath, "rootfs")
	err = DownloadRootFS(distroCfg.URL, downloadName)
	if err != nil {
		return nil, err
	}

	// desempacar
	err = util.UntarRootFS(downloadName, canonicalRootfsPath)
	if err != nil {
		return nil, err
	}
	// montar
	// otro mas
	return &RootFSInstallationConfig{
		InstallationPath:    installationPath,
		Config:              &distroCfg,
		DownloadName:        downloadName,
		Distro:              distroType,
		CanonicalRootfsPath: canonicalRootfsPath,
	}, nil
}

func (c *RootFSInstallationConfig) Mount() {

}
func MountBasics(rootfs string) error {
	mounts := []MountPoint{
		{Src: "proc", Dst: filepath.Join(rootfs, "proc"), Fstype: "proc", Flags: 0},
		{Src: "sysfs", Dst: filepath.Join(rootfs, "sys"), Fstype: "sysfs", Flags: 0},
		{Src: "tmpfs", Dst: filepath.Join(rootfs, "tmp"), Fstype: "tmpfs", Flags: 0},
		{Src: "dev", Dst: filepath.Join(rootfs, "dev"), Fstype: "devtmpfs", Flags: 0},
	}

	for _, m := range mounts {
		os.MkdirAll(m.Dst, 0755)

		if err := unix.Mount(m.Src, m.Dst, m.Fstype, m.Flags, ""); err != nil {
			return fmt.Errorf("failed mount %s at %s: %w", m.Src, m.Dst, err)
		}
	}
	return nil
}
func (c *RootFSInstallationConfig) EnterChroot() error {
	if err := syscall.Chroot(c.CanonicalRootfsPath); err != nil {
		return fmt.Errorf("error en chroot: %w", err)
	}
	return os.Chdir("/")
}

func (c *RootFSInstallationConfig) ExecShell() error {
	return syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
}
