package client

import (
	"axolotl/pkg/command"
	"axolotl/pkg/model"
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
func (c *Client) Create(ctx context.Context, req *model.RunRequest) (*Response, error) {
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

	// Setup stdout and stderr capture
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	// Get stdin pipe to send payload
	stdin, err := session.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	// Channel to track stdin write completion
	stdinDone := make(chan error, 1)

	// Start the command
	cmdString := string(cmd)
	log.Printf("Executing command: %s", cmdString)

	if err := session.Start(cmdString); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}

	// Write payload to stdin in a goroutine
	go func() {
		defer stdin.Close()
		if len(payload) > 0 {
			if _, err := stdin.Write(payload); err != nil {
				stdinDone <- fmt.Errorf("failed to write payload: %w", err)
				return
			}
		}
		stdinDone <- nil
	}()

	// Wait for command completion with context support
	sessionDone := make(chan error, 1)
	go func() {
		sessionDone <- session.Wait()
	}()

	// Wait for either context cancellation or command completion
	var sessionErr error
	select {
	case <-ctx.Done():
		// Context cancelled, try to signal the session
		_ = session.Signal(ssh.SIGTERM)
		// Wait a bit for graceful termination
		select {
		case sessionErr = <-sessionDone:
		case <-make(chan struct{}):
			_ = session.Signal(ssh.SIGKILL)
			sessionErr = <-sessionDone
		}
		return nil, fmt.Errorf("command cancelled: %w", ctx.Err())

	case sessionErr = <-sessionDone:
		// Command completed
	}

	// Check stdin write result
	if err := <-stdinDone; err != nil {
		return nil, err
	}

	// Prepare response
	response := &Response{
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: sessionErr == nil,
	}

	// Extract exit code if available
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
