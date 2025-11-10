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
			log.Printf("❌ Error al aceptar canal: %v", err)
			continue
		}

		go handleChannelRequests(channel, requests)
	}
}

func handleChannelRequests(ch ssh.Channel, reqs <-chan *ssh.Request) {
	defer func() {
		log.Println("🔚 Cerrando canal SSH")
		if err := ch.Close(); err != nil {
			log.Printf("⚠️ Error al cerrar canal: %v", err)
		}
	}()

	for req := range reqs {
		log.Printf("📦 Request: tipo=%s, payload=%x", req.Type, req.Payload)

		switch req.Type {
		case "exec":
			if handleExecRequest(ch, req) {
				return
			}
		default:
			if err := req.Reply(false, nil); err != nil {
				log.Printf("⚠️ Error al responder request: %v", err)
			}
		}
	}
}

func handleExecRequest(ch ssh.Channel, req *ssh.Request) bool {
	var args struct{ Command string }
	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("❌ Error al decodificar comando exec: %v", err)
		if err := req.Reply(false, nil); err != nil {
			log.Printf("⚠️ Error al responder: %v", err)
		}
		return false
	}

	switch args.Command {
	case string(command.AxoRun):
		return handleAxorunCommand(ch, req)
	default:
		log.Printf("⚠️ Comando desconocido: %s", args.Command)
		if err := req.Reply(false, nil); err != nil {
			log.Printf("⚠️ Error al responder: %v", err)
		}
		return false
	}
}
