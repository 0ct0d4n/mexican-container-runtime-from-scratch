package rootfs

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
)

type RootFSEntry struct {
	Arch string
	URL  string
}

var rootfsCatalog = map[string][]RootFSEntry{
	"alpine": {
		{Arch: "arm64", URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/aarch64/alpine-minirootfs-3.20.0-aarch64.tar.gz"},
		{Arch: "amd64", URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz"},
	},

	"ubuntu": {
		{Arch: "arm64", URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-arm64-root.tar.gz"},
		{Arch: "amd64", URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-amd64-root.tar.gz"},
	},

	"debian": {
		{Arch: "arm64", URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-arm64-root.tar.xz"},
		{Arch: "amd64", URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-amd64-root.tar.xz"},
	},

	"busybox": {
		{Arch: "arm64", URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-arm64.tar.gz"},
		{Arch: "amd64", URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-x86_64.tar.gz"},
	},

	"arch": {
		{Arch: "arm64", URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-aarch64.tar.gz"},
		{Arch: "amd64", URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-x86_64.tar.gz"},
	},

	"centos": {
		{Arch: "arm64", URL: "https://cloud.centos.org/centos/9-stream/aarch64/images/CentOS-Stream-GenericCloud-9-20240212.0.aarch64.qcow2"},
		{Arch: "amd64", URL: "https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-9-20240212.0.x86_64.qcow2"},
	},

	"fedora": {
		{Arch: "arm64", URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/aarch64/images/Fedora-Cloud-Base-40-1.15.aarch64.raw.xz"},
		{Arch: "amd64", URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/x86_64/images/Fedora-Cloud-Base-40-1.15.x86_64.raw.xz"},
	},

	"kali": {
		{Arch: "arm64", URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-arm64-rootfs.tar.xz"},
		{Arch: "amd64", URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-amd64-rootfs.tar.xz"},
	},

	"void": {
		{Arch: "arm64", URL: "https://repo-default.voidlinux.org/live/current/void-aarch64-ROOTFS-20240315.tar.xz"},
		{Arch: "amd64", URL: "https://repo-default.voidlinux.org/live/current/void-x86_64-ROOTFS-20240315.tar.xz"},
	},

	"opensuse": {
		{Arch: "arm64", URL: "https://download.opensuse.org/ports/aarch64/tumbleweed/appliances/openSUSE-Tumbleweed-aarch64-RootFS.tar.xz"},
		{Arch: "amd64", URL: "https://download.opensuse.org/tumbleweed/appliances/openSUSE-Tumbleweed-x86_64-RootFS.tar.xz"},
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
	if arch == "" {
		arch = runtime.GOARCH
	}

	arch = normalizeArch(arch)

	entries, ok := rootfsCatalog[distro]
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
