package network

import (
	"fmt"
	"os/exec"
)

// BusyboxNetTool uses /bin/busybox to run "ip" commands inside the container.
// Example commands:
//
//	/bin/busybox ip link set lo up
//	/bin/busybox ip addr add 10.0.0.2/24 dev vethc-xxxx
//	/bin/busybox ip route add default via 10.0.0.1
type BusyboxNetTool struct {
	Path string // e.g. "/bin/busybox"
}

// ensureBusybox returns the correct busybox path or an error if empty.
func ensureBusybox(path string) string {
	if path == "" {
		return "/bin/busybox" // fallback
	}
	return path
}

// run executes a busybox ip command inside the container's namespace.
// The equivalent busybox invocation is:
//
//	busybox ip <args...>
func (b *BusyboxNetTool) run(args ...string) error {
	busybox := ensureBusybox(b.Path)

	cmd := exec.Command(busybox, append([]string{"ip"}, args...)...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("busybox ip failed: %s %v → %w", busybox, args, err)
	}
	return nil
}

// LinkSetUp brings an interface up.
// Example: ip link set eth0 up
func (b *BusyboxNetTool) LinkSetUp(iface string) error {
	return b.run("link", "set", iface, "up")
}

// AddrAdd assigns an IP address to an interface.
// Example: ip addr add 10.0.0.2/24 dev eth0
func (b *BusyboxNetTool) AddrAdd(iface, cidr string) error {
	return b.run("addr", "add", cidr, "dev", iface)
}

// RouteAddDefault sets the default route.
// Example: ip route add default via 10.0.0.1
func (b *BusyboxNetTool) RouteAddDefault(gateway string) error {
	return b.run("route", "add", "default", "via", gateway)
}
