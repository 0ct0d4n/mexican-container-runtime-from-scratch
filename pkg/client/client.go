package client

import (
	"encoding/json"
	"fmt"
	"log"
)

func (r *Client) Create() error {
	data, _ := json.Marshal(r.Req)
	return r.SendPayload(data, "axorun")
}

func (r *Client) SendPayload(data []byte, command string) error {
	stdin, err := r.Session.StdinPipe()
	if err != nil {
		return fmt.Errorf("stdin error: %w", err)
	}

	// ✅ Mandamos el JSON mientras se ejecuta el comando
	go func() {
		defer stdin.Close()
		stdin.Write(data)
	}()

	log.Println("Sending command:", command)
	if err := r.Session.Run(command); err != nil {
		return fmt.Errorf("run error: %w", err)
	}

	return nil
}
