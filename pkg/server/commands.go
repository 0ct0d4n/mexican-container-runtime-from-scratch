package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
	"log"
)

func decodePayloadFromChannel(ch ssh.Channel) (*model.RunRequest, error) {
	var runRequest model.RunRequest
	decoder := json.NewDecoder(ch)

	if err := decoder.Decode(&runRequest); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &runRequest, nil
}

// sendMessage writes a message to the SSH channel.
func sendMessage(ch ssh.Channel, msg string) error {
	if _, err := ch.Write([]byte(msg)); err != nil {
		log.Printf("Error: failed to write to SSH channel: %v", err)
		return err
	}
	return nil
}
