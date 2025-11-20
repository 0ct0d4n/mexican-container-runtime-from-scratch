package network

import (
	"os"
)

type NetTool interface {
	LinkSetUp(iface string) error
	AddrAdd(iface, cidr string) error
	RouteAddDefault(gateway string) error
}

func ResolveNetworkTool(rootfsPath string) NetTool {
	if Exists(rootfsPath + "/bin/busybox") {
		return &BusyboxNetTool{}
	}
	if Exists(rootfsPath + "/sbin/ip") {
		return &IpRoute2Tool{Path: "/sbin/ip"}
	}
	if Exists(rootfsPath + "/bin/ip") {
		return &IpRoute2Tool{Path: "/bin/ip"}
	}
	if Exists(rootfsPath + "/usr/bin/ip") {
		return &IpRoute2Tool{Path: "/usr/bin/ip"}
	}

	//return &DirectSysfsTool{} // ultra fallback
	return nil
}

// Exists returns true if the file or directory exists.
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true // existe
	}
	if os.IsNotExist(err) {
		return false // no existe
	}
	// cualquier otro error (permisos, etc.) lo consideramos como "sí existe"
	return true
}
