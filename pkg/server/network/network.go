package network

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func DebugNETNS() {
	netnsPath := "/proc/self/ns/net"
	target, err := os.Readlink(netnsPath)
	if err != nil {
		log.Printf("[INIT] PID=%d (failed to read %s: %v)", os.Getpid(), netnsPath, err)
	} else {
		log.Printf("[INIT] PID=%d NETNS=%s -> %s", os.Getpid(), netnsPath, target)
	}
}

func VethName(id string) (string, string) {
	s := id[:11]
	return "veth-" + s, "vethc-" + s
}

func CreateVeth(id, pid string) error {
	veth, vethc := VethName(id)
	cmd := exec.Command("ip", "link", "add", veth, "type", "veth", "peer", "name", vethc)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create veth pair: %w", err)
	} else {
		log.Printf("[INIT] Created veth pair veth:%s,vethc:%s", veth, vethc)
	}

	cmd = exec.Command("ip", "link", "set", vethc, "netns", pid)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set vethc into the container: %w", err)
	} else {
		log.Printf("[INIT] successfully set vethc %s into container with PID %s", vethc, pid)
	}

	cmd = exec.Command("ip", "link", "set", veth, "up")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to bring up veth: %w", err)
	} else {
		log.Printf("[INIT] veth %s is up now", vethc)
	}

	return nil
}
