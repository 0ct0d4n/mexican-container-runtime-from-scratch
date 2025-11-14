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
	"strings"
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

	log.Println("about to resolve distro instance for ", config.ImageName)
	distroType := rootfs.DistroType(strings.ToUpper(config.ImageName))
	if distroCfg, err = rootfs.DetectHostDistroConfig(distroType); err != nil {
		log.Println("error detecting host distro config:", err)
		return nil, errors.New("error detecting host distro config")
	}
	installationPath := filepath.Join(diskPath, config.ImageName)
	downloadName := filepath.Join(installationPath, FilenameFromURL(distroCfg.URL))
	canonicalRootfsPath := filepath.Join(installationPath, "rootfs")
	err = DownloadRootFS(distroCfg.URL, downloadName)
	if err != nil {
		return nil, err
	}

	err = util.UntarRootFS(downloadName, canonicalRootfsPath)
	if err != nil {
		return nil, err
	}

	return &RootFSInstallationConfig{
		InstallationPath:    installationPath,
		Config:              &distroCfg,
		DownloadName:        downloadName,
		Distro:              distroType,
		CanonicalRootfsPath: canonicalRootfsPath,
	}, nil
}

func (c *RootFSInstallationConfig) Mount() error {
	err := c.mountBasics()
	if err != nil {
		return err
	}
	err = c.enterChroot()
	if err != nil {
		return err
	}
	err = c.execShell()
	if err != nil {
		return err
	}
	return nil
}

func (c *RootFSInstallationConfig) mountBasics() error {

	// 1. Unshare mount namespace (igual que Docker)
	if err := unix.Unshare(unix.CLONE_NEWNS); err != nil {
		return fmt.Errorf("failed to unshare mount namespace: %w", err)
	}

	// 2. Make it private (prevent mount events from propagating to host)
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}

	mounts := []MountPoint{
		{Src: "proc", Dst: filepath.Join(c.CanonicalRootfsPath, "proc"), Fstype: "proc", Flags: 0},
		{Src: "sysfs", Dst: filepath.Join(c.CanonicalRootfsPath, "sys"), Fstype: "sysfs", Flags: 0},
		{Src: "tmpfs", Dst: filepath.Join(c.CanonicalRootfsPath, "tmp"), Fstype: "tmpfs", Flags: 0},
		{Src: "dev", Dst: filepath.Join(c.CanonicalRootfsPath, "dev"), Fstype: "devtmpfs", Flags: 0},
	}
	for _, m := range mounts {
		os.MkdirAll(m.Dst, 0755)
		if err := unix.Mount(m.Src, m.Dst, m.Fstype, m.Flags, ""); err != nil {
			return fmt.Errorf("failed mount %s at %s: %w", m.Src, m.Dst, err)
		}
	}
	return nil
}
func (c *RootFSInstallationConfig) enterChroot() error {
	if err := syscall.Chroot(c.CanonicalRootfsPath); err != nil {
		return fmt.Errorf("error en chroot: %w", err)
	}
	return os.Chdir("/")
}

func (c *RootFSInstallationConfig) execShell() error {
	return syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
}
