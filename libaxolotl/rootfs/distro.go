// Package rootfs provides root filesystem management for containers
package rootfs

import (
	"fmt"
	"runtime"
	"strings"
)

// DistroType represents a supported Linux distribution
type DistroType string

const (
	Alpine   DistroType = "ALPINE"
	Ubuntu   DistroType = "UBUNTU"
	Debian   DistroType = "DEBIAN"
	Busybox  DistroType = "BUSYBOX"
	Arch     DistroType = "ARCH"
	CentOS   DistroType = "CENTOS"
	Fedora   DistroType = "FEDORA"
	Kali     DistroType = "KALI"
	Void     DistroType = "VOID"
	OpenSUSE DistroType = "OPENSUSE"
)

// FilePurpose describes the purpose of a file in the rootfs
type FilePurpose string

const (
	PurposeShell          FilePurpose = "SHELL"
	PurposeReleaseInfo    FilePurpose = "RELEASE_INFO"
	PurposeInitBinary     FilePurpose = "INIT_BINARY"
	PurposeSystemMetadata FilePurpose = "SYSTEM_METADATA"
)

// FileEntry represents an important file in the rootfs
type FileEntry struct {
	Path     string
	Checksum string
	Purpose  FilePurpose
}

// Config holds rootfs configuration for a specific distro/arch combination
type Config struct {
	URL      string
	Shebangs []FileEntry
}

// Catalog is the global catalog of supported rootfs images
var Catalog = map[DistroType]map[string]Config{
	Alpine: {
		"arm64": {
			URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/aarch64/alpine-minirootfs-3.20.0-aarch64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/sh", Checksum: "abc123", Purpose: PurposeShell},
				{Path: "/etc/alpine-release", Checksum: "def456", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/sh", Checksum: "abc123", Purpose: PurposeShell},
				{Path: "/etc/alpine-release", Checksum: "def456", Purpose: PurposeReleaseInfo},
			},
		},
	},

	Ubuntu: {
		"arm64": {
			URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-arm64-root.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "ghi789", Purpose: PurposeShell},
				{Path: "/etc/lsb-release", Checksum: "jkl012", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-amd64-root.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "ghi789", Purpose: PurposeShell},
				{Path: "/etc/lsb-release", Checksum: "jkl012", Purpose: PurposeReleaseInfo},
			},
		},
	},

	Debian: {
		"arm64": {
			URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-arm64-root.tar.xz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "mno345", Purpose: PurposeShell},
				{Path: "/etc/debian_version", Checksum: "pqr678", Purpose: PurposeSystemMetadata},
			},
		},
		"amd64": {
			URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-amd64-root.tar.xz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "mno345", Purpose: PurposeShell},
				{Path: "/etc/debian_version", Checksum: "pqr678", Purpose: PurposeSystemMetadata},
			},
		},
	},

	Busybox: {
		"arm64": {
			URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-arm64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/busybox", Checksum: "stu901", Purpose: PurposeInitBinary},
			},
		},
		"amd64": {
			URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-x86_64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/busybox", Checksum: "stu901", Purpose: PurposeInitBinary},
			},
		},
	},

	Arch: {
		"arm64": {
			URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-aarch64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "vwx234", Purpose: PurposeShell},
				{Path: "/etc/arch-release", Checksum: "yz0123", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-x86_64.tar.gz",
			Shebangs: []FileEntry{
				{Path: "/bin/bash", Checksum: "vwx234", Purpose: PurposeShell},
				{Path: "/etc/arch-release", Checksum: "yz0123", Purpose: PurposeReleaseInfo},
			},
		},
	},
}

// GetConfig returns the rootfs configuration for a given distro and architecture
func GetConfig(distro DistroType, arch string) (Config, bool) {
	if arch == "" {
		arch = runtime.GOARCH
	}
	archMap, ok := Catalog[distro]
	if !ok {
		return Config{}, false
	}
	cfg, ok := archMap[arch]
	return cfg, ok
}

// NormalizeArch normalizes architecture names to Go's standard names
func NormalizeArch(arch string) string {
	arch = strings.ToLower(arch)

	switch arch {
	case "", "native", "host":
		return runtime.GOARCH
	case "aarch64", "armv8":
		return "arm64"
	case "x86_64", "x64":
		return "amd64"
	case "i386", "i686":
		return "386"
	}

	return arch
}

// HostArch returns the normalized host architecture
func HostArch() string {
	return NormalizeArch(runtime.GOARCH)
}

// DetectDistroConfig returns the rootfs configuration for the given distro on the host architecture
func DetectDistroConfig(distro DistroType) (Config, error) {
	arch := HostArch()

	archMap, ok := Catalog[distro]
	if !ok {
		return Config{}, fmt.Errorf("unsupported distro: %s", distro)
	}

	cfg, ok := archMap[arch]
	if !ok {
		return Config{}, fmt.Errorf(
			"distro %s does not support architecture %s. Supported: %v",
			distro, arch, availableArchs(distro),
		)
	}

	return cfg, nil
}

// ResolveConfig resolves the rootfs configuration for a specific distro and architecture
func ResolveConfig(distro DistroType, arch string) (Config, error) {
	normArch := NormalizeArch(arch)

	cfg, ok := GetConfig(distro, normArch)
	if !ok {
		// Does the distro exist?
		_, distroExists := Catalog[distro]
		if !distroExists {
			return Config{}, fmt.Errorf("unsupported distro: %s", distro)
		}

		// The problem is the architecture
		return Config{}, fmt.Errorf(
			"architecture '%s' not supported for distro %s. Use: %v",
			normArch, distro, availableArchs(distro),
		)
	}

	return cfg, nil
}

func availableArchs(distro DistroType) []string {
	archs := []string{}
	archMap := Catalog[distro]
	for k := range archMap {
		archs = append(archs, k)
	}
	return archs
}
