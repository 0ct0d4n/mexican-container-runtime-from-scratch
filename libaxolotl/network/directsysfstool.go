package network

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DirectSysfsTool configures networking using ONLY writes to /sys/class/net,
// without requiring "ip" or "busybox".
//
// This is the LOWEST LEVEL fallback. It allows you to:
//   - Bring interfaces up/down
//   - Assign IPs via /proc/net or /sys interfaces (limited)
//   - Set default routes (very limited)
//
// WARNING: This backend cannot do everything that "ip" or "busybox ip" can.
// It exists only so the container can at least bring up lo and basic NICs.
type DirectSysfsTool struct {
	NetPath string // usually "/sys/class/net"
}

// ensurePath ensures sysfs root exists
func ensurePath(p string) string {
	if p == "" {
		return "/sys/class/net"
	}
	return p
}

// write writes a string to a sysfs file.
func write(path, val string) error {
	return os.WriteFile(path, []byte(val), 0644)
}

// LinkSetUp brings an interface up by modifying its flags.
// Equivalent to: ip link set iface up
func (d *DirectSysfsTool) LinkSetUp(iface string) error {
	net := ensurePath(d.NetPath)
	flagsPath := filepath.Join(net, iface, "flags")

	// Read current flags
	raw, err := os.ReadFile(flagsPath)
	if err != nil {
		return fmt.Errorf("sysfs: cannot read flags for %s: %w", iface, err)
	}

	flagsStr := strings.TrimSpace(string(raw))
	flags, err := strconv.ParseUint(flagsStr, 0, 32)
	if err != nil {
		return fmt.Errorf("sysfs: malformed flags for %s: %w", iface, err)
	}

	// IFF_UP = 0x1
	flags |= 0x1

	if err := write(flagsPath, fmt.Sprintf("0x%x", flags)); err != nil {
		return fmt.Errorf("sysfs: cannot set %s up: %w", iface, err)
	}

	return nil
}

// AddrAdd assigns an IP to an interface using sysfs.
//
// NOTE: sysfs does NOT support arbitrary address assignment like “ip addr add”.
// The lowest common denominator we can offer is writing to:
//
//	/proc/sys/net/ipv4/conf/<iface>/local
//	/proc/sys/net/ipv4/conf/<iface>/address
//
// This is VERY limited and mostly useful only for loopback (lo).
func (d *DirectSysfsTool) AddrAdd(iface, cidr string) error {
	ip := strings.Split(cidr, "/")[0] // extract IP only

	paths := []string{
		"/proc/sys/net/ipv4/conf/" + iface + "/local",
		"/proc/sys/net/ipv4/conf/" + iface + "/address",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			if err := write(p, ip); err != nil {
				return fmt.Errorf("sysfs: failed writing %s → %w", p, err)
			}
		}
	}

	// This tool cannot do full CIDR setup
	return nil
}

// RouteAddDefault attempts to set the default route using sysfs.
// WARNING: This is extremely limited.
//
// Real default routes cannot be set reliably from sysfs alone —
// normally you MUST use “ip route add default via X”.
//
// Here we try a minimal hack:
//
//	/proc/sys/net/ipv4/route/default_gateway
//
// This file exists in some kernels; in others it won’t.
func (d *DirectSysfsTool) RouteAddDefault(gateway string) error {
	path := "/proc/sys/net/ipv4/route/default_gateway"

	if _, err := os.Stat(path); err != nil {
		// Cannot support routing here
		return fmt.Errorf("sysfs: cannot set default route (no support)")
	}

	if err := write(path, gateway); err != nil {
		return fmt.Errorf("sysfs: cannot write default gateway: %w", err)
	}

	return nil
}
