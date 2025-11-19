package network

import (
	"log"
	"os"
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
