//go:build linux
// +build linux

package libaxolotl

import (
	"axolotl/libaxolotl/network"
	"axolotl/libaxolotl/rootfs"
	"axolotl/libaxolotl/types"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

// InitConfig holds the configuration for container initialization
type InitConfig struct {
	RootfsPath  string
	VethPair    *network.VethPair
	ContainerIP string
	Gateway     string
	Commands    types.Commands
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

// InitContainer performs all container initialization steps
// This function is called from within the container's namespaces (PID=1)
func InitContainer(cfg *InitConfig) error {
	log.Printf("[INIT] Starting container initialization (PID=%d)", os.Getpid())

	// Step 1: Set mount propagation to private
	if err := rootfs.SetMountPropagation(); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}

	// Step 2: Mount basic filesystems (proc, sys, dev, tmp)
	mounts := rootfs.GetDefaultMounts(cfg.RootfsPath)
	if err := rootfs.MountAll(mounts); err != nil {
		return fmt.Errorf("failed to mount filesystems: %w", err)
	}

	// Step 3: Pivot root to the new rootfs
	if err := rootfs.PivotRoot(cfg.RootfsPath); err != nil {
		return fmt.Errorf("failed to pivot root: %w", err)
	}

	// Step 4: Start zombie reaper in background
	go reapZombies()

	log.Printf("[INIT] Container mounted successfully (PID=%d)", os.Getpid())

	// Step 5: Configure container networking

	if err := cfg.ConfigureContainerSide(cfg.ContainerIP); err != nil {
		return fmt.Errorf("failed to configure container network: %w", err)
	}

	// Step 6: Start the main process (like tini)
	return startMainProcess(cfg.Commands.Command, cfg.Commands.Args...)
}

func (initCfg *InitConfig) ConfigureContainerSide(ip string) error {
	log.Printf("[NET] Starting container-side network configuration")
	log.Printf("[NET] Using tool: %T", network.ResolveNetworkTool(initCfg.RootfsPath))

	tool := network.ResolveNetworkTool(initCfg.RootfsPath)

	// --- 1. Validar formato de IP/CIDR ---
	if _, _, err := net.ParseCIDR(ip); err != nil {
		return fmt.Errorf("invalid IP format '%s': %w", ip, err)
	}

	// --- 2. Loopback ---
	if err := tool.LinkSetUp("lo"); err != nil {
		return fmt.Errorf("lo up failed: %w", err)
	}
	log.Printf("[NET] Loopback interface up")

	// --- 3. Levantar interfaz veth del contenedor ---
	if err := tool.LinkSetUp(initCfg.VethPair.ContainerVeth); err != nil {
		return fmt.Errorf("failed to bring up %s: %w", initCfg.VethPair.ContainerVeth, err)
	}
	log.Printf("[NET] Interface %s is up", initCfg.VethPair.ContainerVeth)

	// --- 4. Asignar IP ---
	if err := tool.AddrAdd(initCfg.VethPair.ContainerVeth, ip); err != nil {
		return fmt.Errorf("failed to assign IP %s to %s: %w",
			ip, initCfg.VethPair.ContainerVeth, err)
	}
	log.Printf("[NET] IP %s assigned to %s", ip, initCfg.VethPair.ContainerVeth)

	// --- 5. Agregar default route opcional ---
	gw := "10.0.0.1"
	if initCfg.Gateway != "" {
		gw = initCfg.Gateway
	}

	if err := tool.RouteAddDefault(gw); err != nil {
		log.Printf("[WARN] Could not set default route via %s: %v", gw, err)
	} else {
		log.Printf("[NET] Default route via %s configured", gw)
	}

	log.Printf("[NET] Container-side network configuration complete")

	return nil
}

// startMainProcess starts the main container process and forwards signals (tini-like)
func startMainProcess(command string, args ...string) error {
	log.Printf("[INIT] Launching main process: %s %v", command, args)

	// Prepare child command
	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start child process
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start child process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[INIT] Child started with PID: %d", childPID)

	// Forward signals to child (tini behavior)
	go forwardSignals(childPID)

	// Wait for child to exit
	err := cmd.Wait()

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ProcessState.ExitCode()
		} else {
			log.Printf("[INIT] Error waiting for child: %v", err)
			exitCode = 1
		}
	}

	log.Printf("[INIT] Child exited with code %d", exitCode)

	// Exit init with the same exit code (Docker behavior)
	os.Exit(exitCode)
	return nil
}

// forwardSignals forwards all signals to the child process (tini behavior)
func forwardSignals(childPID int) {
	sigs := make(chan os.Signal, 32)

	// Listen to all signals, like tini
	signal.Notify(sigs)

	for sig := range sigs {
		log.Printf("[INIT] Forwarding signal %v to child PID %d", sig, childPID)
		syscall.Kill(childPID, sig.(syscall.Signal))
	}
}

// reapZombies continuously reaps zombie processes (PID 1 responsibility)
func reapZombies() {
	log.Println("[INIT] Starting zombie reaper...")

	for {
		var status syscall.WaitStatus
		var rusage syscall.Rusage

		// Non-blocking wait for any child process
		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, &rusage)

		if pid > 0 {
			log.Printf("[INIT] Reaped zombie PID=%d status=%d", pid, status.ExitStatus())
			continue
		}

		if err != nil && !errors.Is(err, syscall.ECHILD) {
			log.Printf("[INIT] wait4 error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

// spawnContainer spawns the container process with namespaces
func (initCfg *InitConfig) spawnContainer() error {

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
	if err := network.SetupHostNetworking(initCfg.VethPair, strconv.Itoa(childPID)); err != nil {
		return fmt.Errorf("failed to setup host networking: %w", err)
	}
	log.Printf("[CONTAINER] Host networking configured: %s <-> %s", initCfg.VethPair.HostVeth, initCfg.VethPair.ContainerVeth)

	// Step 7: Wait for container to finish
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container exited with error: %w", err)
	}

	return nil
}
