package server

import (
	"axolotl/pkg/model"
	"encoding/json"
	"golang.org/x/crypto/ssh"
	"log"
)

func handleAxoRunCommand(ch ssh.Channel, req *ssh.Request) bool {
	log.Println("[AXO_RUN] Handler started")

	// Read the JSON payload from STDIN (channel)
	runRequest, err := decodePayloadFromChannel(ch)
	if err != nil {
		log.Printf("Error: failed to read JSON payload: %v", err)
		_ = sendMessage(ch, "Error: invalid payload\n")
		_ = req.Reply(false, nil)
		return true
	}

	if runRequest.Namespace == nil {
		log.Printf("Error: namespace config is nil in request")
		_ = sendMessage(ch, "Error: empty namespace configuration\n")
		_ = req.Reply(false, nil)
		return true
	}

	payload := runRequest.Namespace
	log.Printf("Payload decoded successfully: %+v", payload)

	// Accept the exec request now
	if err := req.Reply(true, nil); err != nil {
		log.Printf("Error: failed to reply to exec-request: %v", err)
		return true
	}

	// Send response to client
	_ = sendMessage(ch, "Processing configuration...\n")

	// Create cgroup
	tmp, err := createCgroup(payload.Cgroup)
	if err != nil {
		log.Printf("[CGROUP] Error: failed to create instance: %v", err)
		_ = sendMessage(ch, "[CGROUP] Error: failed to create instance\n")
		return true
	}
	log.Printf("[CGROUP] Starting configuration at %s", tmp.path)

	// CPU
	if err := tmp.writeCPUMax(); err != nil {
		log.Printf("[CGROUP] Error: failed to configure CPU: %v", err)
		_ = sendMessage(ch, "[CGROUP] Error: failed to configure CPU\n")
		return true
	}

	// Memory
	if err := tmp.writeMemoryMax(); err != nil {
		log.Printf("[CGROUP] Error: failed to configure memory: %v", err)
		_ = sendMessage(ch, "[CGROUP] Error: failed to configure memory\n")
		return true
	}

	// PIDs
	if err := tmp.writePidsMax(); err != nil {
		log.Printf("[CGROUP] Error: failed to configure PIDs: %v", err)
		_ = sendMessage(ch, "[CGROUP] Error: failed to configure PIDs\n")
		return true
	}

	_ = sendMessage(ch, "Configuration completed successfully\n")
	log.Println("[CGROUP] Resources configured successfully")

	return true
}

// decodePayloadFromChannel reads the JSON from the SSH channel (stdin from client).
// The client sends a complete RunRequest with the "namespaces" field.
func decodePayloadFromChannel(ch ssh.Channel) (*model.RunRequest, error) {
	var runRequest model.RunRequest

	decoder := json.NewDecoder(ch)
	if err := decoder.Decode(&runRequest); err != nil {
		log.Printf("Error: failed to read JSON from stdin: %v", err)
		return nil, err
	}

	log.Printf("RunRequest decoded: Command=%v, Namespace=%+v", runRequest.Command, runRequest.Namespace)
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
