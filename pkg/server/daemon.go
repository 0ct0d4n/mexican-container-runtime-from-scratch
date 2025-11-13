package server

import (
	"axolotl/pkg/command"
	"golang.org/x/crypto/ssh"
	"log"
)

func HandleConnection(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {

		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "only sessions allowed")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Error: failed to accept SSH channel: %v", err)
			continue
		}

		go handleChannelRequests(channel, requests)
	}
}

func handleChannelRequests(ch ssh.Channel, reqs <-chan *ssh.Request) {

	for req := range reqs {
		log.Printf("Request received: type=%s payload=%x", req.Type, req.Payload)

		switch req.Type {

		case "exec":
			// If handleExecRequest returns true, we exit
			if handleExecRequest(ch, req) {
				log.Println("Channel completed, closing...")
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

	// Read the command from the exec request
	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("Error: failed to decode exec command: %v", err)
		_ = req.Reply(false, nil)
		return false
	}

	log.Printf("Exec command received: %s", args.Command)

	switch args.Command {

	case string(command.AxoRun):
		return handleAxoRunCommand(ch, req)

	default:
		log.Printf("Warning: unknown command: %s", args.Command)
		_ = req.Reply(false, nil)
		return false
	}
}
