package network

import (
	"fmt"
	"os/exec"
)

type BusyboxNetTool struct {
	Path string
}

func ensureBusybox(path string) string {
	if path == "" {
		return "/bin/busybox"
	}
	return path
}

func (b *BusyboxNetTool) run(args ...string) error {
	busybox := ensureBusybox(b.Path)
	cmd := exec.Command(busybox, append([]string{"ip"}, args...)...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("busybox ip command failed: %s %v: %w", busybox, args, err)
	}
	return nil
}

func (b *BusyboxNetTool) LinkSetUp(iface string) error {
	return b.run("link", "set", iface, "up")
}

func (b *BusyboxNetTool) AddrAdd(iface, cidr string) error {
	return b.run("addr", "add", cidr, "dev", iface)
}

func (b *BusyboxNetTool) RouteAddDefault(gateway string) error {
	return b.run("route", "add", "default", "via", gateway)
}
