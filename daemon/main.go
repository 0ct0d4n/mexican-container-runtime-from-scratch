package main

import (
	axolotl "axolotl/pkg/server"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func main() {
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
-----BEGIN RSA PRIVATE KEY-----
MIIBOgIBAAJBAKoowXQny6ocB/16RMnR7O0uGhSTPZBVzxKGVaKZehD2blJAA5bE
zGLKuqFHo3ZhnV3Xv2hyYyZuzav5Rjh1KPUCAwEAAQJAVAwagAOL9djEqTv5D+Zq
VWYqYQpoWo/3DB8Z3ppdXIMOP5f7apDAflQMj5hw3lFff8f9hU2MEcb5dvzIHoA8
AQIhANbC3u7Mc4WghhYoJhzQx7zPD6DLqYrVDlGSAB7ef4DfAiEAwbhySgWROF1K
lJj6PMaF2E8ik6ABYxKMZ2J1FZC4pAkCIFwSxR8x0ekShTj9q+v5XrH+KxlG4FtE
j0dcnlGGwVgHAiBnmJXwKy1r3ZJfPhA3EJ4aCQeoJZ3G6GfKTvVFuJ8cvQIgAUXm
RVXz69Zy6ThG9ZblAnH1q/1ko2nNa6z5wZ/4EG8=
-----END RSA PRIVATE KEY-----
`
