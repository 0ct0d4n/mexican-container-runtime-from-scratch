package client

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

// Config holds the SSH client configuration.
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	PrivateKeyPath  string
	HostKeyPath     string
	Timeout         time.Duration
	InsecureHostKey bool
}

// DefaultConfig returns a Config with default values.
func DefaultConfig() *Config {
	return &Config{
		Host:            "localhost",
		Port:            2222,
		User:            "axolotl",
		Password:        "",
		Timeout:         30 * time.Second,
		InsecureHostKey: false,
	}
}

// LoadConfigFromEnv loads configuration from environment variables.
// Falls back to defaults if not set.
func LoadConfigFromEnv() *Config {
	cfg := DefaultConfig()

	if host := os.Getenv("AXOLOTL_HOST"); host != "" {
		cfg.Host = host
	}

	if port := os.Getenv("AXOLOTL_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if user := os.Getenv("AXOLOTL_USER"); user != "" {
		cfg.User = user
	}

	if password := os.Getenv("AXOLOTL_PASSWORD"); password != "" {
		cfg.Password = password
	}

	if keyPath := os.Getenv("AXOLOTL_PRIVATE_KEY"); keyPath != "" {
		cfg.PrivateKeyPath = keyPath
	}

	if hostKeyPath := os.Getenv("AXOLOTL_HOST_KEY"); hostKeyPath != "" {
		cfg.HostKeyPath = hostKeyPath
	}

	if insecure := os.Getenv("AXOLOTL_INSECURE_HOST_KEY"); insecure == "true" {
		cfg.InsecureHostKey = true
	}

	if timeout := os.Getenv("AXOLOTL_TIMEOUT"); timeout != "" {
		if d, err := time.ParseDuration(timeout); err == nil {
			cfg.Timeout = d
		}
	}

	return cfg
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}

	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port: %d", c.Port)
	}

	if c.User == "" {
		return fmt.Errorf("user cannot be empty")
	}

	if c.Password == "" && c.PrivateKeyPath == "" {
		return fmt.Errorf("either password or private key must be provided")
	}

	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}

	return nil
}

// SSHClientConfig creates an ssh.ClientConfig from the Config.
func (c *Config) SSHClientConfig() (*ssh.ClientConfig, error) {
	config := &ssh.ClientConfig{
		User:    c.User,
		Timeout: c.Timeout,
	}

	// Setup authentication
	var authMethods []ssh.AuthMethod

	if c.Password != "" {
		authMethods = append(authMethods, ssh.Password(c.Password))
	}

	if c.PrivateKeyPath != "" {
		key, err := os.ReadFile(c.PrivateKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read private key: %w", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}

		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication method available")
	}

	config.Auth = authMethods

	// Setup host key callback
	if c.InsecureHostKey {
		config.HostKeyCallback = ssh.InsecureIgnoreHostKey()
	} else if c.HostKeyPath != "" {
		callback, err := hostKeyCallbackFromFile(c.HostKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to setup host key verification: %w", err)
		}
		config.HostKeyCallback = callback
	} else {
		return nil, fmt.Errorf("host key verification required: set HostKeyPath or enable InsecureHostKey")
	}

	return config, nil
}

// Address returns the formatted address string (host:port).
func (c *Config) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// hostKeyCallbackFromFile creates a HostKeyCallback from a known_hosts file.
func hostKeyCallbackFromFile(path string) (ssh.HostKeyCallback, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	hostKey, _, _, _, err := ssh.ParseAuthorizedKey(data)
	if err != nil {
		return nil, err
	}

	return ssh.FixedHostKey(hostKey), nil
}
