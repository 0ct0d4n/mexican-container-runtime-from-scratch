package network

import (
	"fmt"
	"os/exec"
)

type IpRoute2Tool struct {
	Path string
}

func ensureIP(path string) string {
	if path == "" {
		return "/sbin/ip"
	}
	return path
}

func (t *IpRoute2Tool) run(args ...string) error {
	ip := ensureIP(t.Path)
	cmd := exec.Command(ip, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ip command failed: %s %v: %w", ip, args, err)
	}
	return nil
}

func (t *IpRoute2Tool) LinkSetUp(iface string) error {
	return t.run("link", "set", iface, "up")
}

func (t *IpRoute2Tool) AddrAdd(iface, cidr string) error {
	return t.run("addr", "add", cidr, "dev", iface)
}

func (t *IpRoute2Tool) RouteAddDefault(gateway string) error {
	return t.run("route", "add", "default", "via", gateway)
}
