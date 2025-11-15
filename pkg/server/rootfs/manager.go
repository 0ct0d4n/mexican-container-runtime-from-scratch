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
func CleanupMounts(c *RootFSInstallationConfig) {
	log.Println("about to unmount points for ", c.CanonicalRootfsPath)
	dirs := GetMountPoint(c)
	for _, d := range dirs {
		syscall.Unmount(d.Dst, 0)
		syscall.Unmount(d.Dst, syscall.MNT_DETACH)
	}
}

func (c *RootFSInstallationConfig) Mount() error {
	if err := unix.Unshare(
		unix.CLONE_NEWNS |
			unix.CLONE_NEWPID |
			unix.CLONE_NEWUTS |
			unix.CLONE_NEWIPC,
	); err != nil {
		return fmt.Errorf("failed to unshare namespaces: %w", err)
	}

	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}

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
	log.Println("Container mounted successfully :D enjoy!")
	return nil
}

func (c *RootFSInstallationConfig) mountBasics() error {
	mounts := GetMountPoint(c)
	for _, m := range mounts {
		err := os.MkdirAll(m.Dst, 0755)
		if err != nil {
			return err
		}
		if err := unix.Mount(m.Src, m.Dst, m.Fstype, m.Flags, ""); err != nil {
			return fmt.Errorf("failed mount %s at %s: %w", m.Src, m.Dst, err)
		}
	}
	return nil
}

func GetMountPoint(c *RootFSInstallationConfig) []MountPoint {
	mounts := []MountPoint{
		{Src: "proc", Dst: filepath.Join(c.CanonicalRootfsPath, "proc"), Fstype: "proc", Flags: 0},
		{Src: "sysfs", Dst: filepath.Join(c.CanonicalRootfsPath, "sys"), Fstype: "sysfs", Flags: 0},
		{Src: "tmpfs", Dst: filepath.Join(c.CanonicalRootfsPath, "tmp"), Fstype: "tmpfs", Flags: 0},
		{Src: "dev", Dst: filepath.Join(c.CanonicalRootfsPath, "dev"), Fstype: "devtmpfs", Flags: 0},
	}
	return mounts
}

func (c *RootFSInstallationConfig) enterChroot() error {
	if err := syscall.Chroot(c.CanonicalRootfsPath); err != nil {
		return fmt.Errorf("error en chroot: %w %v", err, c.CanonicalRootfsPath)
	}
	return os.Chdir("/")
}

func (c *RootFSInstallationConfig) execShell() error {
	return syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", "echo CONTAINER_IS_WORKING_NOW!; "}, os.Environ())
}
