package main

import (
	axolotl "axolotl/pkg/server"
	"golang.org/x/crypto/ssh"
	"log"
	"net"
)

func startDaemonMode() {
	log.Println("Starting daemon")
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	private, err := ssh.ParsePrivateKey([]byte(testPrivateKey))
	if err != nil {
		log.Fatal(err)
	}
	config.AddHostKey(private)

	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		log.Fatal("Failed to listen:", err)
	}
	log.Println("🦎 Axolotl SSH Daemon escuchando en puerto 2222...")

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
			axolotl.HandleConnection(chans)
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
