// Package daemon provides SSH daemon functionality for the axolotl runtime
package daemon

import (
	"axolotl/libaxolotl"
	"axolotl/libaxolotl/types"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

// HandleConnection handles incoming SSH connections
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
			// Handle exec request and close channel when done
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

	// Parse SSH exec command
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

		// Decode payload from channel
		payload, err := decodePayload(ch)
		if err != nil {
			log.Printf("Error decoding payload: %v", err)
			return true
		}

		// Start container
		err = libaxolotl.Start(payload)
		if err != nil {
			log.Printf("[CONTAINER] Error: %v", err)
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

func decodePayload(ch ssh.Channel) (*types.RunRequest, error) {
	var runRequest types.RunRequest
	decoder := json.NewDecoder(ch)

	if err := decoder.Decode(&runRequest); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &runRequest, nil
}
