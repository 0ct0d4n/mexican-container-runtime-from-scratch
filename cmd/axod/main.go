package main

import (
	"axolotl/internal/daemon"
	"axolotl/libaxolotl"
	"axolotl/libaxolotl/cgroups"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "init-container" {
		log.Println("[AXOD] Running as init-container (PID 1)")
		// Run the container initialization sequence
		if err := runContainerInit(); err != nil {
			log.Printf("[AXOD] Init-container failed: %v", err)
			os.Exit(1)
		}
		log.Println("[AXOD] Init-container started successfully")
		return
	}

	// Otherwise this is the daemon process
	startDaemonMode()
}

func runContainerInit() error {
	var initCfg *libaxolotl.InitConfig
	if err := json.NewDecoder(os.Stdin).Decode(&initCfg); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// Execute container initialization
	return libaxolotl.InitContainer(initCfg)
}

func startDaemonMode() {
	log.Println("[AXOD] Starting Axolotl daemon")

	// Ensure axolotl cgroup root exists with controllers enabled
	if err := cgroups.EnsureAxolotlRoot(); err != nil {
		log.Printf("[AXOD] Warning: cgroup root setup: %v (may already exist)", err)
	}

	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	private, err := ssh.ParsePrivateKey([]byte(testPrivateKey))
	if err != nil {
		log.Fatal("[AXOD] Failed to parse private key:", err)
	}
	config.AddHostKey(private)

	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal("[AXOD] Failed to listen on port 2222:", err)
	}
	log.Println("[AXOD] SSH daemon listening on port 2222")

	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Println("failed to accept incoming connection:", err)
			continue
		}

		go func(c net.Conn) {
			sshConn, chans, reqs, err := ssh.NewServerConn(c, config)
			if err != nil {
				log.Println("failed to handshake:", err)
				return
			}
			defer sshConn.Close()

			go ssh.DiscardRequests(reqs)
			daemon.HandleConnection(chans)
		}(tcpConn)
	}
}

const testPrivateKey = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBoqqvL36g/Y7hKAs+fAv8fMy8JS1dxT7IbLaPKivpRVwAAAJBe4bbVXuG2
1QAAAAtzc2gtZWQyNTUxOQAAACBoqqvL36g/Y7hKAs+fAv8fMy8JS1dxT7IbLaPKivpRVw
AAAEAG2Nw6bCLbmWPAMI+sHl+4dzRi/L7ibaMakyXMEB7vh2iqq8vfqD9juEoCz58C/x8z
LwlLV3FPshsto8qK+lFXAAAADWF4b2xvdGxAbG9jYWw=
-----END OPENSSH PRIVATE KEY-----
`
