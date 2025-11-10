package server

import (
	"axolotl/pkg/model"
	"golang.org/x/crypto/ssh"
	"log"
)

func HandleConnection(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "solo sesiones")
			continue
		}

		channel, requests, _ := newChannel.Accept()
		go func() {
			for req := range requests {
				log.Println("Received request:", string(req.Payload))
				switch req.Type {
				case "exec":
					var payload model.NamespaceConfig
					err := ssh.Unmarshal(req.Payload, &payload)
					if err != nil {
						return
					}
					log.Println(payload)
					//cmd := exec.Command("bash", "-c ", payload.Command)
					//cmd.Stdout = channel
					//cmd.Stderr = channel
					//cmd.Run()
					channel.Close()
				default:
					req.Reply(false, nil)
				}
			}
		}()
	}
}
