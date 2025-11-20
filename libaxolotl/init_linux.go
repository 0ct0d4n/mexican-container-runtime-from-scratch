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

type InitConfig struct {
	RootfsPath  string
	VethPair    *network.VethPair
	ContainerIP string
	Gateway     string
	Commands    types.Commands
}

func createInitConfig(container *rootfs.Container) *InitConfig {
	return &InitConfig{
		RootfsPath:  container.RootfsPath,
		ContainerIP: "10.0.0.2/24",
		Commands:    container.Commands,
		VethPair:    network.NewVethPair(container.ID),
	}
}

// InitContainer performs container initialization from within PID namespace (PID=1).
// This is called by the child process after clone with namespaces.
func InitContainer(cfg *InitConfig) error {
	log.Printf("[INIT] Container initialization started (PID=%d)", os.Getpid())

	if err := rootfs.SetMountPropagation(); err != nil {
		return fmt.Errorf("failed to set mount propagation: %w", err)
	}

	mounts := rootfs.GetDefaultMounts(cfg.RootfsPath)
	if err := rootfs.MountAll(mounts); err != nil {
		return fmt.Errorf("failed to mount filesystems: %w", err)
	}

	if err := rootfs.PivotRoot(cfg.RootfsPath); err != nil {
		return fmt.Errorf("failed to pivot root: %w", err)
	}

	go reapZombies()

	log.Printf("[INIT] Filesystem ready, configuring network")

	if err := cfg.ConfigureContainerSide(cfg.ContainerIP); err != nil {
		return fmt.Errorf("failed to configure container network: %w", err)
	}

	return startMainProcess(cfg.Commands.Command, cfg.Commands.Args...)
}

func (initCfg *InitConfig) ConfigureContainerSide(ip string) error {
	tool := network.ResolveNetworkTool(initCfg.RootfsPath)
	log.Printf("[NET] Configuring network using %T", tool)

	if _, _, err := net.ParseCIDR(ip); err != nil {
		return fmt.Errorf("invalid IP format '%s': %w", ip, err)
	}

	if err := tool.LinkSetUp("lo"); err != nil {
		return fmt.Errorf("failed to bring up loopback: %w", err)
	}

	if err := tool.LinkSetUp(initCfg.VethPair.ContainerVeth); err != nil {
		return fmt.Errorf("failed to bring up %s: %w", initCfg.VethPair.ContainerVeth, err)
	}

	if err := tool.AddrAdd(initCfg.VethPair.ContainerVeth, ip); err != nil {
		return fmt.Errorf("failed to assign IP %s to %s: %w",
			ip, initCfg.VethPair.ContainerVeth, err)
	}
	log.Printf("[NET] IP %s assigned to %s", ip, initCfg.VethPair.ContainerVeth)

	gw := "10.0.0.1"
	if initCfg.Gateway != "" {
		gw = initCfg.Gateway
	}

	if err := tool.RouteAddDefault(gw); err != nil {
		log.Printf("[NET] Warning: could not set default route via %s: %v", gw, err)
	} else {
		log.Printf("[NET] Default route configured via %s", gw)
	}

	return nil
}

// startMainProcess launches the main container process with signal forwarding (tini-like behavior)
func startMainProcess(command string, args ...string) error {
	log.Printf("[INIT] Starting process: %s %v", command, args)

	cmd := exec.Command(command, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[INIT] Process started with PID %d", childPID)

	go forwardSignals(childPID)

	err := cmd.Wait()
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ProcessState.ExitCode()
		} else {
			log.Printf("[INIT] Process wait error: %v", err)
			exitCode = 1
		}
	}

	log.Printf("[INIT] Process exited with code %d", exitCode)
	os.Exit(exitCode)
	return nil
}

func forwardSignals(childPID int) {
	sigs := make(chan os.Signal, 32)
	signal.Notify(sigs)

	for sig := range sigs {
		log.Printf("[INIT] Forwarding signal %v to PID %d", sig, childPID)
		syscall.Kill(childPID, sig.(syscall.Signal))
	}
}

func reapZombies() {
	for {
		var status syscall.WaitStatus
		var rusage syscall.Rusage

		pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, &rusage)

		if pid > 0 {
			log.Printf("[INIT] Reaped zombie process PID %d (exit status %d)", pid, status.ExitStatus())
			continue
		}

		if err != nil && !errors.Is(err, syscall.ECHILD) {
			log.Printf("[INIT] Zombie reaper error: %v", err)
		}

		time.Sleep(100 * time.Millisecond)
	}
}

func (initCfg *InitConfig) spawnContainer() error {
	cmd := exec.Command("/proc/self/exe", "init-container")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,
	}

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

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start container process: %w", err)
	}

	childPID := cmd.Process.Pid
	log.Printf("[CONTAINER] Container process started (PID %d)", childPID)

	if err := network.SetupHostNetworking(initCfg.VethPair, strconv.Itoa(childPID)); err != nil {
		return fmt.Errorf("failed to setup host networking: %w", err)
	}
	log.Printf("[CONTAINER] Network configured: %s <-> %s", initCfg.VethPair.HostVeth, initCfg.VethPair.ContainerVeth)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("container exited with error: %w", err)
	}

	return nil
}
