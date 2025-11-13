package rootfs

import (
	"fmt"
	"strings"
)

type DistroType string

const (
	ALPINE   DistroType = "ALPINE"
	UBUNTU   DistroType = "UBUNTU"
	DEBIAN   DistroType = "DEBIAN"
	BUSYBOX  DistroType = "BUSYBOX"
	ARCH     DistroType = "ARCH"
	CENTOS   DistroType = "CENTOS"
	FEDORA   DistroType = "FEDORA"
	KALI     DistroType = "KALI"
	VOID     DistroType = "VOID"
	OPENSUSE DistroType = "OPENSUSE"
)

type RootFSEntry struct {
	Path     string
	Checksum string
}

var rootfsCatalog = map[DistroType][]RootFSEntry{
	ALPINE: {
		{Path: "/bin/sh", Checksum: "abc123"},
		{Path: "/etc/alpine-release", Checksum: "def456"},
	},
	UBUNTU: {
		{Path: "/bin/bash", Checksum: "ghi789"},
		{Path: "/etc/lsb-release", Checksum: "jkl012"},
	},
	DEBIAN: {
		{Path: "/bin/bash", Checksum: "mno345"},
		{Path: "/etc/debian_version", Checksum: "pqr678"},
	},
	BUSYBOX: {
		{Path: "/bin/busybox", Checksum: "stu901"},
	},
	ARCH: {
		{Path: "/bin/bash", Checksum: "vwx234"},
		{Path: "/etc/arch-release", Checksum: "yz0123"},
	},
	CENTOS: {
		{Path: "/bin/bash", Checksum: "abc456"},
		{Path: "/etc/centos-release", Checksum: "def789"},
	},
	FEDORA: {
		{Path: "/bin/bash", Checksum: "ghi012"},
		{Path: "/etc/fedora-release", Checksum: "jkl345"},
	},
	KALI: {
		{Path: "/bin/bash", Checksum: "mno678"},
		{Path: "/etc/kali-version", Checksum: "pqr901"},
	},
	VOID: {
		{Path: "/bin/bash", Checksum: "stu234"},
		{Path: "/etc/void-release", Checksum: "vwx567"},
	},
	OPENSUSE: {
		{Path: "/bin/bash", Checksum: "yz8901"},
		{Path: "/etc/SuSE-release", Checksum: "abc234"},
	},
}

func GetRootFSURL(distro string) (string, error) {
	distro = strings.ToLower(distro)
	distroKey := DistroType(strings.ToUpper(distro))
	entries, ok := rootfsCatalog[distroKey]
	if !ok {
		return "", fmt.Errorf("unsupported distro: %s", distro)
	}
	// dummy logic to form URL from entries
	return fmt.Sprintf("https://example.com/rootfs/%s.tar.gz", distro), nil
}
