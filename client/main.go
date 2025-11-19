package main

import (
	"axolotl/pkg/model"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	axolotl "axolotl/pkg/client"
)

func main() {
	// Parse command line flags
	var (
		host          = flag.String("host", "", "SSH server host (default: from env or localhost)")
		port          = flag.Int("port", 0, "SSH server port (default: from env or 2222)")
		user          = flag.String("user", "", "SSH username (default: from env or axolotl)")
		password      = flag.String("password", "", "SSH password (default: from env)")
		keyFile       = flag.String("key", "", "Path to private key file (default: from env)")
		insecure      = flag.Bool("insecure", false, "Skip host key verification (INSECURE)")
		timeout       = flag.Duration("timeout", 0, "Connection timeout (default: from env or 30s)")
		containerName = flag.String("name", "test2", "Container name")
		imageName     = flag.String("image", "alpine", "image name")
		memoryMB      = flag.Int64("memory", 100, "Memory limit in MB")
		cpuPercent    = flag.Float64("cpu", 0.5, "CPU limit (0.0-1.0)")
		maxPids       = flag.Uint64("pids", 100, "Maximum number of PIDs")
		cgroupPath    = flag.String("cgroup-path", "test", "Cgroup path")
	)
	flag.Parse()

	// Load configuration from environment variables
	config := axolotl.LoadConfigFromEnv()

	// Override with command line flags if provided
	if *host != "" {
		config.Host = *host
	}
	if *port != 0 {
		config.Port = *port
	}
	if *user != "" {
		config.User = *user
	}
	if *password != "" {
		config.Password = *password
	}
	if *keyFile != "" {
		config.PrivateKeyPath = *keyFile
	}
	if *insecure {
		config.InsecureHostKey = true
	}
	if *timeout != 0 {
		config.Timeout = *timeout
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	log.Printf("Connecting to %s as %s...", config.Address(), config.User)

	// Create client
	client, err := axolotl.NewClient(config)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Warning: failed to close client: %v", err)
		}
	}()

	log.Println("Connected successfully")

	// Setup context with cancellation on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupt signals gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nReceived interrupt signal, cancelling...")
		cancel()
	}()

	// Build request using the builder pattern
	request, err := axolotl.NewRequestBuilder().
		WithContainerName(*containerName).
		WithOrg("global").
		WithImage(*imageName).
		WithCommand(model.Commands{
			Command: "/bin/sh",
			Args:    []string{"-c", " /bin/busybox  ip addr add 10.0.0.2/24 dev veth-cont && ip link set veth-cont up && ip link set lo up   "},
		}).
		WithCgroup(*cgroupPath, *memoryMB, *cpuPercent, *maxPids).
		Build()
	if err != nil {
		log.Fatalf("Failed to build request: %v", err)
	}

	log.Printf("Creating container '%s' with limits: memory=%dMB, cpu=%.1f%%, pids=%d",
		*containerName, *memoryMB, *cpuPercent*100, *maxPids)

	// Execute the request
	response, err := client.Create(ctx, request)
	if err != nil {
		log.Fatalf("Failed to create container: %v", err)
	}

	// Display the response
	fmt.Println("\n" + response.String())

	// Check if the operation was successful
	if !response.Success {
		log.Fatalf("Container creation failed: %v", response.Error())
	}

	log.Println("Container created successfully!")
}

// Example usage with environment variables:
//
// export AXOLOTL_HOST=192.168.64.2
// export AXOLOTL_PORT=2222
// export AXOLOTL_USER=axolotl
// export AXOLOTL_PASSWORD=axolotl
// export AXOLOTL_INSECURE_HOST_KEY=true
// export AXOLOTL_TIMEOUT=30s
//
// ./client
//
// Or with command line flags:
//
// ./client -host 192.168.64.2 -user axolotl -password axolotl -insecure -memory 200 -cpu 0.8 -image alpine
