package network

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

type VethPair struct {
	HostVeth      string
	ContainerVeth string
}

func NewVethPair(containerID string) *VethPair {
	shortID := containerID
	if len(shortID) > 7 {
		shortID = shortID[:7]
	}

	return &VethPair{
		HostVeth:      "veth-" + shortID,
		ContainerVeth: "vethc-" + shortID,
	}
}

func (v *VethPair) Create() error {
	cmd := exec.Command("ip", "link", "add", v.HostVeth, "type", "veth", "peer", "name", v.ContainerVeth)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	}

	log.Printf("[NET] Created veth pair: %s <-> %s", v.HostVeth, v.ContainerVeth)
	return nil
}

func (v *VethPair) MoveToNamespace(pid string) error {
	cmd := exec.Command("ip", "link", "set", v.ContainerVeth, "netns", pid)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to move %s to netns %s: %w", v.ContainerVeth, pid, err)
	}

	log.Printf("[NET] Moved %s to container netns (PID %s)", v.ContainerVeth, pid)
	return nil
}

func (v *VethPair) BringUpHost() error {
	cmd := exec.Command("ip", "link", "set", v.HostVeth, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up %s: %w", v.HostVeth, err)
	}

	log.Printf("[NET] Host interface %s is up", v.HostVeth)
	return nil
}

func SetupHostNetworking(veth *VethPair, containerPID string) error {
	if err := veth.Create(); err != nil {
		return err
	}

	if err := veth.MoveToNamespace(containerPID); err != nil {
		return err
	}

	if err := veth.BringUpHost(); err != nil {
		return err
	}

	return nil
}
