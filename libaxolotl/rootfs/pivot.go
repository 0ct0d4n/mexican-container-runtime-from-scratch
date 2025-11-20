//go:build linux
// +build linux

package rootfs

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
)

// PivotRoot performs the pivot_root system call to change the root filesystem
// This is more secure than chroot as it properly isolates the container
func PivotRoot(newRoot string) error {
	// Step 1: Bind mount newRoot onto itself to make it a mount point
	if err := unix.Mount(newRoot, newRoot, "", unix.MS_BIND|unix.MS_REC, ""); err != nil {
		return fmt.Errorf("failed to bind-mount rootfs on itself (%s): %w", newRoot, err)
	}

	// Step 2: Create putOld directory for old root
	putOld := filepath.Join(newRoot, ".pivot_root_old")
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("failed to create putOld dir %s: %w", putOld, err)
	}

	// Step 3: Call pivot_root
	if err := unix.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root(%s, %s) failed: %w", newRoot, putOld, err)
	}

	// Step 4: Change to the new root
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("failed to chdir(\"/\"): %w", err)
	}

	// Step 5: Unmount and remove old root
	oldRoot := "/.pivot_root_old"
	if err := unix.Unmount(oldRoot, unix.MNT_DETACH); err != nil {
		return fmt.Errorf("failed to unmount old root (%s): %w", oldRoot, err)
	}
	if err := os.RemoveAll(oldRoot); err != nil {
		return fmt.Errorf("failed to remove old root dir (%s): %w", oldRoot, err)
	}

	return nil
}
