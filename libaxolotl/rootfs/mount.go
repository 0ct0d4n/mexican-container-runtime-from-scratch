//go:build linux
// +build linux

package rootfs

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

type MountPoint struct {
	Source string
	Target string
	FSType string
	Flags  uintptr
	Data   string
}

func GetDefaultMounts(rootfsPath string) []MountPoint {
	return []MountPoint{
		{
			Source: "proc",
			Target: filepath.Join(rootfsPath, "proc"),
			FSType: "proc",
			Flags:  0,
		},
		{
			Source: "sysfs",
			Target: filepath.Join(rootfsPath, "sys"),
			FSType: "sysfs",
			Flags:  0,
		},
		{
			Source: "tmpfs",
			Target: filepath.Join(rootfsPath, "tmp"),
			FSType: "tmpfs",
			Flags:  0,
		},
		{
			Source: "dev",
			Target: filepath.Join(rootfsPath, "dev"),
			FSType: "devtmpfs",
			Flags:  0,
		},
	}
}

func MountAll(mounts []MountPoint) error {
	for _, m := range mounts {
		if err := os.MkdirAll(m.Target, 0755); err != nil {
			return fmt.Errorf("failed to create mount point %s: %w", m.Target, err)
		}

		if err := unix.Mount(m.Source, m.Target, m.FSType, m.Flags, ""); err != nil {
			return fmt.Errorf("failed to mount %s at %s: %w", m.Source, m.Target, err)
		}
	}
	return nil
}

func UnmountAll(mounts []MountPoint) {
	for _, m := range mounts {
		unix.Unmount(m.Target, 0)
		unix.Unmount(m.Target, unix.MNT_DETACH)
	}
}

func SetMountPropagation() error {
	if err := unix.Mount("", "/", "", unix.MS_REC|unix.MS_PRIVATE, ""); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}
	return nil
}
