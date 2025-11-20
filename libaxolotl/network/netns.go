package network

import (
	"log"
	"os"
)

// DebugNamespace logs the current network namespace information
// Useful for debugging network isolation
func DebugNamespace() {
	netnsPath := "/proc/self/ns/net"
	target, err := os.Readlink(netnsPath)
	if err != nil {
		log.Printf("[NETNS] PID=%d (failed to read %s: %v)", os.Getpid(), netnsPath, err)
	} else {
		log.Printf("[NETNS] PID=%d NETNS=%s -> %s", os.Getpid(), netnsPath, target)
	}
}
