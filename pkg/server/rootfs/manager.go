//go:build linux
// +build linux

package rootfs

import (
	"axolotl/pkg/model"
	"axolotl/pkg/util"
	"axolotl/pkg/util/tini"
	"errors"
	"fmt"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
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
	Commands            []model.Commands
}

func InstallImage(diskPath string, config *model.NamespaceConfig, command []model.Commands) (*RootFSInstallationConfig, error) {
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
		Commands:            command,
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
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}

	err := c.mountBasics()
	if err != nil {
		return err
	}
	err = c.pivotRoot()
	if err != nil {
		return err
	}
	go c.reapZombies()
	log.Println("Container mounted successfully :D enjoy! current PID=", os.Getpid())

	for _, command := range c.Commands {
		err := tini.StartMainProcess(command.Command, command.Args...)
		if err != nil {
			return err
		}
	}
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

// @deprecated it's unsecure please do not use.
func (c *RootFSInstallationConfig) enterChroot() error {
	if err := syscall.Chroot(c.CanonicalRootfsPath); err != nil {
		return fmt.Errorf("error en chroot: %w %v", err, c.CanonicalRootfsPath)
	}
	return os.Chdir("/")
}
func (c *RootFSInstallationConfig) pivotRoot() error {
	rootfs := c.CanonicalRootfsPath

	if err := unix.Mount(rootfs, rootfs, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("pivot_root: failed to bind-mount rootfs on itself (%s): %w", rootfs, err)
	}

	putOld := filepath.Join(rootfs, ".pivot_root_old")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("pivot_root: failed to create putOld dir %s: %w", putOld, err)
	}

	if err := unix.PivotRoot(rootfs, putOld); err != nil {
		return fmt.Errorf("pivot_root: unix.PivotRoot(%s, %s) failed: %w", rootfs, putOld, err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("pivot_root: failed to chdir(\"/\"): %w", err)
	}

	oldRoot := "/.pivot_root_old"
	if err := unix.Unmount(oldRoot, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("pivot_root: failed to unmount old root (%s): %w", oldRoot, err)
	}
	if err := os.RemoveAll(oldRoot); err != nil {
		return fmt.Errorf("pivot_root: failed to remove old root dir (%s): %w", oldRoot, err)
	}

	return nil
}
func (c *RootFSInstallationConfig) execShell() error {
	//return syscall.Exec("/bin/sh", []string{"/bin/sh"}, os.Environ())
	log.Println("Container mounted successfully :D enjoy! current PID=", os.Getpid())
	return syscall.Exec("/bin/sh", []string{"/bin/sh", "-c", "echo CONTAINER_IS_WORKING_NOW!; "}, os.Environ())
}

func (c *RootFSInstallationConfig) reapZombies() {
	log.Println("Reaping Zombies...")
	for {
		var status syscall.WaitStatus
		var rusage syscall.Rusage

		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, &rusage)
		if pid > 0 {
			log.Printf("[INIT] Reaped zombie pid=%d status=%d", pid, status.ExitStatus())
			continue
		}

		if err != nil && !errors.Is(err, syscall.ECHILD) {
			log.Printf("[INIT] wait4 error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
