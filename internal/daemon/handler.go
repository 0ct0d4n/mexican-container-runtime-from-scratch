package daemon

import (
	"axolotl/libaxolotl"
	"axolotl/libaxolotl/types"
	"encoding/json"
	"fmt"
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
			log.Printf("[SSH] Failed to accept channel: %v", err)
			continue
		}

		go handleChannelRequests(channel, requests)
	}
}

func handleChannelRequests(ch ssh.Channel, reqs <-chan *ssh.Request) {
	for req := range reqs {
		log.Printf("[SSH] Request: type=%s", req.Type)

		switch req.Type {
		case "exec":
			if handleExecRequest(ch, req) {
				log.Printf("[SSH] Closing channel")
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

	if err := ssh.Unmarshal(req.Payload, &args); err != nil {
		log.Printf("[SSH] Failed to unmarshal exec command: %v", err)
		_ = req.Reply(false, nil)
		return true
	}

	log.Printf("[SSH] Executing command: %s", args.Command)

	switch args.Command {
	case "AXO_RUN":
		if err := req.Reply(true, nil); err != nil {
			log.Printf("[SSH] Failed to reply: %v", err)
			return true
		}

		payload, err := decodePayload(ch)
		if err != nil {
			log.Printf("[SSH] Failed to decode payload: %v", err)
			return true
		}

		err = libaxolotl.Start(payload)
		if err != nil {
			log.Printf("[SSH] Container start failed: %v", err)
			ch.Write([]byte("error\n"))
		} else {
			ch.Write([]byte("ok\n"))
		}

		ch.SendRequest("exit-status", false, ssh.Marshal(struct{ Status uint32 }{0}))
		return true

	default:
		log.Printf("[SSH] Unknown command: %s", args.Command)
		_ = req.Reply(false, nil)
		return true
	}
}

func decodePayload(ch ssh.Channel) (*types.RunRequest, error) {
	var runRequest types.RunRequest
	decoder := json.NewDecoder(ch)

	if err := decoder.Decode(&runRequest); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &runRequest, nil
}
