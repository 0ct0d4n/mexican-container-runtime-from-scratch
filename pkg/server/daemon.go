package server

import (
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

	// 1️⃣ Leer solo el comando SSH ("AXO_RUN")
	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("Error unmarshalling exec command: %v", err)
		_ = req.Reply(false, nil)
		return true
	}

	log.Printf("Exec command received: %s", args.Command)

	switch args.Command {

	case "AXO_RUN":
		if err := req.Reply(true, nil); err != nil {
			log.Printf("Error replying to exec request: %v", err)
			return true
		}

		payload, err := decodePayloadFromChannel(ch)
		if err != nil {
			log.Printf("Error decoding payload: %v", err)
			return true
		}

		err = StartContainer(err, payload)
		if err != nil {
			log.Printf("[CGROUP] Error: failed to create instance: %v", err)
			ch.Write([]byte("error\n"))
		} else {

			ch.Write([]byte("ok\n"))
		}

		ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
		return true

	default:
		log.Printf("Unknown command: %s", args.Command)
		_ = req.Reply(false, nil)
		return true
	}
}
