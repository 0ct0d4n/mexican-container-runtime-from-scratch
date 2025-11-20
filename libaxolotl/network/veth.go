// Package network provides network namespace and veth pair management
package network

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// VethPair represents a virtual ethernet pair for container networking
type VethPair struct {
	HostVeth      string // veth interface on the host
	ContainerVeth string // veth interface inside container
}

// NewVethPair creates a new veth pair with names based on container ID
func NewVethPair(containerID string) *VethPair {
	// Truncate ID to keep interface names short (max 15 chars for ifname)
	shortID := containerID
	if len(shortID) > 9 {
		shortID = shortID[:9]
	}

	return &VethPair{
		HostVeth:      "veth-" + shortID,
		ContainerVeth: "vethc-" + shortID,
	}
}

// Create creates the veth pair on the host
func (v *VethPair) Create() error {
	cmd := exec.Command("ip", "link", "add", v.HostVeth, "type", "veth", "peer", "name", v.ContainerVeth)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}

	log.Printf("[NETWORK] Created veth pair: host=%s container=%s", v.HostVeth, v.ContainerVeth)
	return nil
}

// MoveToNamespace moves the container veth to the specified network namespace (by PID)
func (v *VethPair) MoveToNamespace(pid string) error {
	cmd := exec.Command("ip", "link", "set", v.ContainerVeth, "netns", pid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move %s to netns %s: %w", v.ContainerVeth, pid, err)
	}

	log.Printf("[NETWORK] Moved %s to netns of PID=%s", v.ContainerVeth, pid)
	return nil
}

// BringUpHost brings up the host-side veth interface
func (v *VethPair) BringUpHost() error {
	cmd := exec.Command("ip", "link", "set", v.HostVeth, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up %s: %w", v.HostVeth, err)
	}

	log.Printf("[NETWORK] Host veth %s is UP", v.HostVeth)
	return nil
}

// ConfigureContainerSide configures the container-side networking
// This should be called from inside the container's network namespace
func (v *VethPair) ConfigureContainerSide(ip string) error {
	commands := [][]string{
		{"ip", "link", "set", "lo", "up"},
		{"ip", "link", "set", v.ContainerVeth, "up"},
		{"ip", "addr", "add", ip, "dev", v.ContainerVeth},
	}

	for _, cmdArgs := range commands {
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("container network config failed (%v): %w", cmdArgs, err)
		}
	}

	log.Printf("[NETWORK] Container network configured: %s on %s", ip, v.ContainerVeth)
	return nil
}

// SetupHostNetworking performs all host-side network setup
// 1. Creates veth pair
// 2. Moves container veth to container's netns
// 3. Brings up host veth
func SetupHostNetworking(veth *VethPair, containerPID string) error {

	// Create veth pair
	if err := veth.Create(); err != nil {
		return err
	}

	// Move container veth to container's namespace
	if err := veth.MoveToNamespace(containerPID); err != nil {
		return err
	}

	// Bring up host veth
	if err := veth.BringUpHost(); err != nil {
		return err
	}

	return nil
}
