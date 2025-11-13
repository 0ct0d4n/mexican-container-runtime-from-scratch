package rootfs

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type RootFSEntry struct {
	Arch     string
	URL      string
	Path     string
	Checksum string
}

type DistroType string

const ALPINE DistroType = "ALPINE"
const UBUNTU DistroType = "UBUNTU"
const DEBIAN DistroType = "DEBIAN"
const BUSYBOX DistroType = "BUSYBOX"
const ARCH DistroType = "ARCH"
const CENTOS DistroType = "CENTOS"
const FEDORA DistroType = "FEDORA"
const KALI DistroType = "KALI"
const VOID DistroType = "VOID"
const OPENSUSE DistroType = "OPENSUSE"

var RootfsCatalog = map[DistroType][]RootFSEntry{
	ALPINE: {
		{Arch: "arm64", URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/aarch64/alpine-minirootfs-3.20.0-aarch64.tar.gz"},
		{Arch: "amd64", URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz"},
	},

	UBUNTU: {
		{Arch: "arm64", URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-arm64-root.tar.gz"},
		{Arch: "amd64", URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-amd64-root.tar.gz"},
	},

	DEBIAN: {
		{Arch: "arm64", URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-arm64-root.tar.xz"},
		{Arch: "amd64", URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-amd64-root.tar.xz"},
	},

	BUSYBOX: {
		{Arch: "arm64", URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-arm64.tar.gz"},
		{Arch: "amd64", URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-x86_64.tar.gz"},
	},

	ARCH: {
		{Arch: "arm64", URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-aarch64.tar.gz"},
		{Arch: "amd64", URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-x86_64.tar.gz"},
	},

	CENTOS: {
		{Arch: "arm64", URL: "https://cloud.centos.org/centos/9-stream/aarch64/images/CentOS-Stream-GenericCloud-9-20240212.0.aarch64.qcow2"},
		{Arch: "amd64", URL: "https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-9-20240212.0.x86_64.qcow2"},
	},

	FEDORA: {
		{Arch: "arm64", URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/aarch64/images/Fedora-Cloud-Base-40-1.15.aarch64.raw.xz"},
		{Arch: "amd64", URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/x86_64/images/Fedora-Cloud-Base-40-1.15.x86_64.raw.xz"},
	},

	KALI: {
		{Arch: "arm64", URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-arm64-rootfs.tar.xz"},
		{Arch: "amd64", URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-amd64-rootfs.tar.xz"},
	},

	VOID: {
		{Arch: "arm64", URL: "https://repo-default.voidlinux.org/live/current/void-aarch64-ROOTFS-20240315.tar.xz"},
		{Arch: "amd64", URL: "https://repo-default.voidlinux.org/live/current/void-x86_64-ROOTFS-20240315.tar.xz"},
	},

	OPENSUSE: {
		{Arch: "arm64", URL: "https://download.opensuse.org/ports/aarch64/tumbleweed/appliances/openSUSE-Tumbleweed-aarch64-RootFS.tar.xz"},
		{Arch: "amd64", URL: "https://download.opensuse.org/tumbleweed/appliances/openSUSE-Tumbleweed-x86_64-RootFS.tar.xz"},
	},
}

var RootfsSheBangCatalog = map[DistroType][]RootFSEntry{
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

// normalizeArch converts Go runtime arch -> standard Linux arch names
func normalizeArch(a string) string {
	a = strings.ToLower(a)
	switch a {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return a
	}
}

// GetRootFSURL returns the URL for a given distro + arch.
// If arch = "" → auto-detect using runtime.GOARCH.
func GetRootFSURL(arch, distro string) (string, error) {
	distro = strings.ToLower(distro)
	distroKey := DistroType(strings.ToUpper(distro))
	if arch == "" {
		arch = runtime.GOARCH
	}

	arch = normalizeArch(arch)

	entries, ok := RootfsCatalog[distroKey]
	if !ok {
		return "", fmt.Errorf("unsupported distro: %s", distro)
	}

	for _, e := range entries {
		if e.Arch == arch {
			return e.URL, nil
		}
	}

	return "", errors.New("no URL found for distro " + distro + " and arch " + arch)
}
