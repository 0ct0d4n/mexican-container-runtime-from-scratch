package network

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	// DefaultBridgeName is the default bridge interface name
	DefaultBridgeName = "axol0"
)

// Bridge represents a Linux bridge for container networking
type Bridge struct {
	Name string
	IP   string
}

// NewBridge creates a new bridge configuration
func NewBridge(name, ip string) *Bridge {
	if name == "" {
		name = DefaultBridgeName
	}
	return &Bridge{
		Name: name,
		IP:   ip,
	}
}

// Create creates the bridge interface if it doesn't exist
func (b *Bridge) Create() error {
	// Check if bridge already exists
	cmd := exec.Command("ip", "link", "show", b.Name)
	if err := cmd.Run(); err == nil {
		log.Printf("[BRIDGE] Bridge %s already exists", b.Name)
		return nil
	}

	// Create bridge
	cmd = exec.Command("ip", "link", "add", b.Name, "type", "bridge")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create bridge %s: %w", b.Name, err)
	}

	log.Printf("[BRIDGE] Created bridge %s", b.Name)
	return nil
}

// BringUp brings up the bridge interface
func (b *Bridge) BringUp() error {
	cmd := exec.Command("ip", "link", "set", b.Name, "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up bridge %s: %w", b.Name, err)
	}

	log.Printf("[BRIDGE] Bridge %s is UP", b.Name)
	return nil
}

// AssignIP assigns an IP address to the bridge
func (b *Bridge) AssignIP() error {
	if b.IP == "" {
		return nil
	}

	cmd := exec.Command("ip", "addr", "add", b.IP, "dev", b.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// Ignore error if address already exists
		log.Printf("[BRIDGE] IP assignment warning: %v (may already exist)", err)
		return nil
	}

	log.Printf("[BRIDGE] Assigned IP %s to bridge %s", b.IP, b.Name)
	return nil
}

// AttachInterface attaches a veth interface to the bridge
func (b *Bridge) AttachInterface(ifname string) error {
	cmd := exec.Command("ip", "link", "set", ifname, "master", b.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to attach %s to bridge %s: %w", ifname, b.Name, err)
	}

	log.Printf("[BRIDGE] Attached %s to bridge %s", ifname, b.Name)
	return nil
}

// Setup performs complete bridge setup (create, bring up, assign IP)
func (b *Bridge) Setup() error {
	if err := b.Create(); err != nil {
		return err
	}
	if err := b.AssignIP(); err != nil {
		return err
	}
	if err := b.BringUp(); err != nil {
		return err
	}
	return nil
}
