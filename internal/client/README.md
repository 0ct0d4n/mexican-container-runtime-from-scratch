# Axolotl Client Library

Professional Go client library for interacting with the Axolotl SSH daemon.

## Features

- ✅ **Configuration Management**: Environment variables, config files, and command-line flags
- ✅ **Type-Safe Builder Pattern**: Fluent API for constructing requests with validation
- ✅ **Context Support**: Proper context handling for timeouts and cancellation
- ✅ **Response Capture**: Captures stdout/stderr from server operations
- ✅ **Error Handling**: Comprehensive error handling with wrapped errors
- ✅ **Connection Management**: Proper resource cleanup and connection pooling support
- ✅ **Thread-Safe**: Safe for concurrent use
- ✅ **Security**: Support for both password and key-based authentication with host key verification

## Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "log"

    "axolotl/pkg/client"
)

func main() {
    // Load configuration from environment
    config := client.LoadConfigFromEnv()
    config.Host = "192.168.64.2"
    config.Port = 2222
    config.User = "axolotl"
    config.Password = "axolotl"
    config.InsecureHostKey = true // Only for development!

    // Create client
    cli, err := client.NewClient(config)
    if err != nil {
        log.Fatal(err)
    }
    defer cli.Close()

    // Build request
    req, err := client.NewRequestBuilder().
        WithContainerName("my-container").
        WithCgroup("test", 100, 0.5, 100).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Execute
    ctx := context.Background()
    resp, err := cli.Create(ctx, req)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Success: %t\n%s", resp.Success, resp.Stdout)
}
```

## Configuration

### Environment Variables

The client supports the following environment variables:

```bash
export AXOLOTL_HOST=192.168.64.2
export AXOLOTL_PORT=2222
export AXOLOTL_USER=axolotl
export AXOLOTL_PASSWORD=axolotl
export AXOLOTL_PRIVATE_KEY=/path/to/key
export AXOLOTL_HOST_KEY=/path/to/known_hosts
export AXOLOTL_INSECURE_HOST_KEY=true  # INSECURE - only for dev
export AXOLOTL_TIMEOUT=30s
```

### Programmatic Configuration

```go
config := &client.Config{
    Host:            "192.168.64.2",
    Port:            2222,
    User:            "axolotl",
    Password:        "secret",
    PrivateKeyPath:  "/path/to/key",
    HostKeyPath:     "/path/to/known_hosts",
    Timeout:         30 * time.Second,
    InsecureHostKey: false,
}

cli, err := client.NewClient(config)
```

## Request Builder

The `RequestBuilder` provides a fluent API for constructing requests with validation:

### Basic Cgroup Configuration

```go
req, err := client.NewRequestBuilder().
    WithContainerName("my-container").
    WithCgroup("path/to/cgroup", 512, 0.8, 200).  // 512MB, 80% CPU, 200 PIDs
    Build()
```

### Advanced Configuration

```go
req, err := client.NewRequestBuilder().
    WithContainerName("advanced-container").
    WithCgroup("test/advanced", 1024, 1.0, 500).
    WithUTS("my-hostname", "example.com").
    WithPID([]string{"/bin/init"}).
    WithNetwork(true, []string{"eth0"}, map[string]string{"default": "192.168.1.1"}).
    WithMount("/rootfs", []string{"/proc", "/sys"}, false).
    WithIPC(64 * 1024 * 1024).  // 64MB shared memory
    WithUser(
        map[int]int{0: 1000},  // UID mapping
        map[int]int{0: 1000},  // GID mapping
    ).
    WithTime(3600).  // 1 hour offset
    Build()
```

### Raw Configuration

For fine-grained control:

```go
cgroupConfig := &model.CgroupNamespace{
    Path:      "custom/path",
    MemoryMax: 2 * units.GB,
    CPUMax:    1.5,
    CPUQuota:  150000,
    CPUPeriod: 100000,
    PidsMax:   1000,
}

req, err := client.NewRequestBuilder().
    WithContainerName("custom").
    WithCgroupRaw(cgroupConfig).
    Build()
```

## Response Handling

The `Response` type provides structured access to command results:

```go
resp, err := cli.Create(ctx, req)
if err != nil {
    log.Fatal(err)
}

// Check success
if !resp.Success {
    log.Printf("Failed: %v", resp.Error())
}

// Access output
fmt.Println(resp.Stdout)
fmt.Println(resp.Stderr)
fmt.Printf("Exit code: %d\n", resp.ExitCode)

// Pretty print
fmt.Println(resp.String())
```

## Context and Cancellation

All operations support context for timeouts and cancellation:

```go
// With timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

resp, err := cli.Create(ctx, req)

// With cancellation
ctx, cancel := context.WithCancel(context.Background())

// Cancel from signal handler
go func() {
    <-signalChan
    cancel()
}()

resp, err := cli.Create(ctx, req)
```

## Error Handling

All errors are wrapped with context using `fmt.Errorf` with `%w`:

```go
resp, err := cli.Create(ctx, req)
if err != nil {
    // Errors include full context
    log.Printf("Operation failed: %v", err)

    // Can unwrap for specific error types
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Operation timed out")
    }
}
```

## Connection Management

### Single Connection

```go
cli, err := client.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer cli.Close()

// Use client...
```

### Multiple Operations

The client is thread-safe and can be reused:

```go
cli, err := client.NewClient(config)
if err != nil {
    log.Fatal(err)
}
defer cli.Close()

// Execute multiple operations
for i := 0; i < 10; i++ {
    req := buildRequest(i)
    resp, err := cli.Create(ctx, req)
    // Handle response...
}
```

### Connection Health Check

```go
ctx := context.Background()
if err := cli.Ping(ctx); err != nil {
    log.Printf("Connection unhealthy: %v", err)
}
```

## Security Best Practices

### Development

For development, you can use insecure mode:

```go
config.InsecureHostKey = true  // Skip host key verification
```

### Production

In production, always verify host keys:

```go
config := &client.Config{
    Host:            "server.example.com",
    Port:            2222,
    User:            "axolotl",
    PrivateKeyPath:  "/secure/path/to/key",
    HostKeyPath:     "/secure/path/to/known_hosts",
    Timeout:         30 * time.Second,
    InsecureHostKey: false,  // REQUIRED in production
}
```

### Authentication

Prefer key-based authentication over passwords:

```go
// Generate key pair
ssh-keygen -t ed25519 -f axolotl_key

// Use in config
config.PrivateKeyPath = "axolotl_key"
```

## Testing

### Mock Client

For testing, you can implement the client interface:

```go
type MockClient struct {
    CreateFunc func(context.Context, *model.RunRequest) (*Response, error)
}

func (m *MockClient) Create(ctx context.Context, req *model.RunRequest) (*Response, error) {
    return m.CreateFunc(ctx, req)
}
```

## Examples

See `client/main.go` for a complete example with command-line flags and signal handling.

### Run with Environment Variables

```bash
export AXOLOTL_HOST=192.168.64.2
export AXOLOTL_USER=axolotl
export AXOLOTL_PASSWORD=axolotl
export AXOLOTL_INSECURE_HOST_KEY=true

./client
```

### Run with Flags

```bash
./client \
  -host 192.168.64.2 \
  -user axolotl \
  -password axolotl \
  -insecure \
  -name my-container \
  -memory 200 \
  -cpu 0.8 \
  -pids 500
```

## License

See main project LICENSE.
