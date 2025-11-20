//go:build linux
// +build linux

package rootfs

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// PivotRoot changes the root filesystem using pivot_root (more secure than chroot)
func PivotRoot(newRoot string) error {
	if err := unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind-mount rootfs: %w", err)
	}

	putOld := filepath.Join(newRoot, ".pivot_root_old")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("failed to create pivot_root temp dir: %w", err)
	}

	if err := unix.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root failed: %w", err)
	}

	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("chdir to new root failed: %w", err)
	}

	oldRoot := "/.pivot_root_old"
	if err := unix.Unmount(oldRoot, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount old root: %w", err)
	}
	if err := os.RemoveAll(oldRoot); err != nil {
		return fmt.Errorf("failed to remove old root: %w", err)
	}

	return nil
}
