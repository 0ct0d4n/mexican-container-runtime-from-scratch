package network

import (
	"fmt"
	"os/exec"
)

// IpRoute2Tool executes the native "ip" command inside the container.
// Typical paths: /sbin/ip, /bin/ip, /usr/bin/ip
//
// Example usage:
//
//	ip link set lo up
//	ip addr add 10.0.0.2/24 dev eth0
//	ip route add default via 10.0.0.1
type IpRoute2Tool struct {
	Path string // full path to the "ip" binary
}

// ensureIP returns a usable ip binary path.
func ensureIP(path string) string {
	if path == "" {
		return "/sbin/ip" // safest fallback
	}
	return path
}

// run executes an ip command inside the container's namespace.
//
// Example:
//
//	run("link", "set", "lo", "up")
//
// → executes:
//
//	/sbin/ip link set lo up
func (t *IpRoute2Tool) run(args ...string) error {
	ip := ensureIP(t.Path)

	cmd := exec.Command(ip, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("iproute2 command failed: %s %v → %w", ip, args, err)
	}
	return nil
}

// LinkSetUp sets the interface up.
// ip link set eth0 up
func (t *IpRoute2Tool) LinkSetUp(iface string) error {
	return t.run("link", "set", iface, "up")
}

// AddrAdd assigns an IP address.
// ip addr add 10.0.0.2/24 dev eth0
func (t *IpRoute2Tool) AddrAdd(iface, cidr string) error {
	return t.run("addr", "add", cidr, "dev", iface)
}

// RouteAddDefault sets the default route.
// ip route add default via 10.0.0.1
func (t *IpRoute2Tool) RouteAddDefault(gateway string) error {
	return t.run("route", "add", "default", "via", gateway)
}
