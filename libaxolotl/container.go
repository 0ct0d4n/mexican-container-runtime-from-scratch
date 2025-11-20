//go:build linux
// +build linux

package libaxolotl

import (
	"axolotl/libaxolotl/cgroups"
	"axolotl/libaxolotl/network"
	"axolotl/libaxolotl/rootfs"
	"axolotl/libaxolotl/types"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"syscall"
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
	if err := spawnContainer(container); err != nil {
		return fmt.Errorf("failed to spawn container: %w", err)
	}

	log.Printf("[CONTAINER] Container %s completed successfully", request.Namespace.ContainerName)
	return nil
}

// spawnContainer spawns the container process with namespaces
func spawnContainer(container *rootfs.Container) error {
	// Prepare init config
	initCfg := &InitConfig{
		RootfsPath:    container.RootfsPath,
		ContainerVeth: container.ContainerVeth,
		ContainerIP:   "10.0.0.2/24",
		Commands:      container.Commands,
	}

	// Re-execute ourselves in init mode with namespaces
	cmd := exec.Command("/proc/self/exe", "init-container")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS | // Mount namespace
			syscall.CLONE_NEWPID | // PID namespace
			syscall.CLONE_NEWUTS | // UTS namespace (hostname)
			syscall.CLONE_NEWIPC | // IPC namespace
			syscall.CLONE_NEWNET, // Network namespace
	}

	// Pass init config via stdin
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	go func() {
		json.NewEncoder(stdin).Encode(initCfg)
		stdin.Close()
	}()

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the child process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[CONTAINER] Child process started with PID=%d", childPID)

	// Step 6: Setup host-side networking (veth pair)
	veth, err := network.SetupHostNetworking(container.ID, strconv.Itoa(childPID))
	if err != nil {
		return fmt.Errorf("failed to setup host networking: %w", err)
	}

	log.Printf("[CONTAINER] Host networking configured: %s <-> %s", veth.HostVeth, veth.ContainerVeth)

	// Step 7: Wait for container to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container exited with error: %w", err)
	}

	return nil
}
