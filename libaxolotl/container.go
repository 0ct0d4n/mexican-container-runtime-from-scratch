//go:build linux
// +build linux

package libaxolotl

import (
	"axolotl/libaxolotl/cgroups"
	"axolotl/libaxolotl/network"
	"axolotl/libaxolotl/rootfs"
	"axolotl/libaxolotl/types"
	"fmt"
	"log"
)

const (
	// DefaultImagesPath is the default path for storing rootfs images
	DefaultImagesPath = "/var/axolotl/images/"
)

// Start creates and starts a new container
func Start(request *types.RunRequest) error {
	log.Printf("[CONTAINER] Starting container: %s", request.Namespace.ContainerName)

	// Step 1: Create cgroup
	cgroupManager, err := cgroups.NewManager(request.Namespace)
	if err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	err = cgroupManager.CreateCgroup()
	if err != nil {
		return err
	}

	// Step 2: Apply resource limits
	if err := cgroupManager.ApplyLimits(); err != nil {
		return fmt.Errorf("failed to apply cgroup limits: %w", err)
	}

	// Step 3: Install rootfs
	container, err := rootfs.Install(DefaultImagesPath, request.Namespace, request.Command)
	if err != nil {
		return fmt.Errorf("failed to install rootfs: %w", err)
	}

	// Step 4: Cleanup any existing mounts
	container.Cleanup()

	log.Printf("[CONTAINER] Rootfs ready, spawning container process")

	// Step 5: Spawn container process
	initCfg := createInitConfig(container)
	if err := initCfg.spawnContainer(); err != nil {
		return fmt.Errorf("failed to spawn container: %w", err)
	}

	log.Printf("[CONTAINER] Container %s completed successfully", request.Namespace.ContainerName)
	return nil
}

// spawnContainer spawns the container process with namespaces
func createInitConfig(container *rootfs.Container) *InitConfig {
	// Prepare init config
	return &InitConfig{
		RootfsPath:  container.RootfsPath,
		ContainerIP: "10.0.0.2/24",
		Commands:    container.Commands,
		VethPair:    network.NewVethPair(container.ID),
	}
}
