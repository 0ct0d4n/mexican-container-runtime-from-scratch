package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/ssh"
)

func decodePayloadFromChannel(ch ssh.Channel) (*model.RunRequest, error) {
	var runRequest model.RunRequest
	decoder := json.NewDecoder(ch)

	if err := decoder.Decode(&runRequest); err != nil {
		return nil, fmt.Errorf("decode error: %w", err)
	}

	return &runRequest, nil
}
