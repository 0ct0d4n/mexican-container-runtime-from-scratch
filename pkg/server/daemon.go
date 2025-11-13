package server

import (
	"axolotl/pkg/command"
	"golang.org/x/crypto/ssh"
	"log"
)

func HandleConnection(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {

		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "solo sesiones")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("❌ Error aceptando canal SSH: %v", err)
			continue
		}

		go handleChannelRequests(channel, requests)
	}
}

func handleChannelRequests(ch ssh.Channel, reqs <-chan *ssh.Request) {

	for req := range reqs {
		log.Printf("📦 Request recibido: tipo=%s payload=%x", req.Type, req.Payload)

		switch req.Type {

		case "exec":
			// Si handleExecRequest devuelve true, salimos
			if handleExecRequest(ch, req) {
				log.Println("🔚 Canal completado, cerrando...")
				_ = ch.Close()
				return
			}

		default:
			req.Reply(false, nil)
		}
	}
}

func handleExecRequest(ch ssh.Channel, req *ssh.Request) bool {
	var args struct{ Command string }

	// 1️⃣ leer el comando del exec request (AXO_RUN)
	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("❌ Error decodificando comando exec: %v", err)
		_ = req.Reply(false, nil)
		return false
	}

	log.Printf("🚀 Exec command recibido: %s", args.Command)

	switch args.Command {

	case string(command.AxoRun): // ← CORRECTO
		return handleAxoRunCommand(ch, req)

	default:
		log.Printf("⚠️ Comando desconocido: %s", args.Command)
		_ = req.Reply(false, nil)
		return false
	}
}
