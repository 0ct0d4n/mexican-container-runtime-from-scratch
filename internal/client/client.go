package client

import (
	"axolotl/internal/command"
	"axolotl/libaxolotl/types"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"

	"golang.org/x/crypto/ssh"
)

// Client represents an Axolotl SSH client.
type Client struct {
	config     *Config
	sshClient  *ssh.Client
	connection io.Closer
	mu         sync.Mutex
}

// NewClient creates a new Axolotl client with the given configuration.
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	sshConfig, err := config.SSHClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH config: %w", err)
	}

	sshClient, err := ssh.Dial("tcp", config.Address(), sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", config.Address(), err)
	}

	return &Client{
		config:     config,
		sshClient:  sshClient,
		connection: sshClient,
	}, nil
}

// Close closes the SSH connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connection != nil {
		err := c.connection.Close()
		c.connection = nil
		return err
	}

	return nil
}

// Create creates a container with the given namespace configuration.
func (c *Client) Create(ctx context.Context, req *types.RunRequest) (*Response, error) {
	if req == nil {
		return nil, fmt.Errorf("request cannot be nil")
	}

	if req.Namespace == nil {
		return nil, fmt.Errorf("namespace configuration cannot be nil")
	}

	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.execute(ctx, command.AxoRun, data)
}

// execute runs a command on the server with the given payload.
func (c *Client) execute(ctx context.Context, cmd command.AxoCommand, payload []byte) (*Response, error) {
	c.mu.Lock()
	if c.sshClient == nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("client is closed")
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		c.mu.Unlock()
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	c.mu.Unlock()
	defer session.Close()

	// Capture stdout & stderr
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Stdin pipe
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// IMPORTANT: start command BEFORE writing JSON
	cmdString := string(cmd)
	log.Printf("Executing command: %s", cmdString)

	if err := session.Start(cmdString); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Write JSON payload
	go func() {
		defer stdin.Close()
		if len(payload) > 0 {
			_, err := stdin.Write(payload)
			if err != nil {
				log.Printf("ERROR writing to stdin: %v", err)
			}
		}
	}()

	// Context-aware Wait()
	sessionDone := make(chan error, 1)
	go func() {
		sessionDone <- session.Wait()
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(ssh.SIGTERM)
		<-sessionDone
		return nil, fmt.Errorf("command cancelled: %w", ctx.Err())

	case sessionErr := <-sessionDone:
		// Build response
		response := &Response{
			Stdout:  stdout.String(),
			Stderr:  stderr.String(),
			Success: sessionErr == nil,
		}

		if exitErr, ok := sessionErr.(*ssh.ExitError); ok {
			response.ExitCode = exitErr.ExitStatus()
		} else if sessionErr == nil {
			response.ExitCode = 0
		} else {
			response.ExitCode = -1
		}

		log.Printf("Command completed: exit_code=%d, success=%t", response.ExitCode, response.Success)
		return response, nil
	}
}

// Ping checks if the connection to the server is alive.
func (c *Client) Ping(ctx context.Context) error {
	c.mu.Lock()
	if c.sshClient == nil {
		c.mu.Unlock()
		return fmt.Errorf("client is closed")
	}

	session, err := c.sshClient.NewSession()
	if err != nil {
		c.mu.Unlock()
		return fmt.Errorf("failed to create session: %w", err)
	}
	c.mu.Unlock()

	defer session.Close()

	// Simple echo command to test connectivity
	done := make(chan error, 1)
	go func() {
		done <- session.Run("echo ping")
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
