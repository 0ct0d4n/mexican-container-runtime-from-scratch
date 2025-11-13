package rootfs

import (
	"fmt"
	"runtime"
	"strings"
)

//
// ──────────────────────────────────────────────────────────────
//   ENUMS
// ──────────────────────────────────────────────────────────────
//

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

type ShebangPurpose string

const (
	PurposeShell          ShebangPurpose = "SHELL"
	PurposeReleaseInfo    ShebangPurpose = "RELEASE_INFO"
	PurposeInitBinary     ShebangPurpose = "INIT_BINARY"
	PurposeSystemMetadata ShebangPurpose = "SYSTEM_METADATA"
)

//
// ──────────────────────────────────────────────────────────────
//   MAIN STRUCTS
// ──────────────────────────────────────────────────────────────
//

type RootFSEntry struct {
	Path     string
	Checksum string
	Purpose  ShebangPurpose
}

type RootFSConfig struct {
	URL      string
	Shebangs []RootFSEntry
}

//
// ──────────────────────────────────────────────────────────────
//   ROOTFS CATALOG
// ──────────────────────────────────────────────────────────────
//

var RootfsCatalog = map[DistroType]map[string]RootFSConfig{

	// ────────────────────────
	//   ALPINE
	// ────────────────────────
	ALPINE: {
		"arm64": {
			URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/aarch64/alpine-minirootfs-3.20.0-aarch64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/sh", Checksum: "abc123", Purpose: PurposeShell},
				{Path: "/etc/alpine-release", Checksum: "def456", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://dl-cdn.alpinelinux.org/alpine/v3.20/releases/x86_64/alpine-minirootfs-3.20.0-x86_64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/sh", Checksum: "abc123", Purpose: PurposeShell},
				{Path: "/etc/alpine-release", Checksum: "def456", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   UBUNTU
	// ────────────────────────
	UBUNTU: {
		"arm64": {
			URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-arm64-root.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "ghi789", Purpose: PurposeShell},
				{Path: "/etc/lsb-release", Checksum: "jkl012", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://partner-images.canonical.com/core/jammy/current/ubuntu-jammy-core-cloudimg-amd64-root.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "ghi789", Purpose: PurposeShell},
				{Path: "/etc/lsb-release", Checksum: "jkl012", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   DEBIAN
	// ────────────────────────
	DEBIAN: {
		"arm64": {
			URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-arm64-root.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "mno345", Purpose: PurposeShell},
				{Path: "/etc/debian_version", Checksum: "pqr678", Purpose: PurposeSystemMetadata},
			},
		},
		"amd64": {
			URL: "https://cloud.debian.org/images/cloud/bullseye/latest/debian-11-genericcloud-amd64-root.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "mno345", Purpose: PurposeShell},
				{Path: "/etc/debian_version", Checksum: "pqr678", Purpose: PurposeSystemMetadata},
			},
		},
	},

	// ────────────────────────
	//   BUSYBOX
	// ────────────────────────
	BUSYBOX: {
		"arm64": {
			URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-arm64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/busybox", Checksum: "stu901", Purpose: PurposeInitBinary},
			},
		},
		"amd64": {
			URL: "https://landley.net/aboriginal/downloads/busybox-rootfs-x86_64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/busybox", Checksum: "stu901", Purpose: PurposeInitBinary},
			},
		},
	},

	// ────────────────────────
	//   ARCH
	// ────────────────────────
	ARCH: {
		"arm64": {
			URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-aarch64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "vwx234", Purpose: PurposeShell},
				{Path: "/etc/arch-release", Checksum: "yz0123", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://geo.mirror.pkgbuild.com/iso/latest/archlinux-bootstrap-latest-x86_64.tar.gz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "vwx234", Purpose: PurposeShell},
				{Path: "/etc/arch-release", Checksum: "yz0123", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   CENTOS
	// ────────────────────────
	CENTOS: {
		"arm64": {
			URL: "https://cloud.centos.org/centos/9-stream/aarch64/images/CentOS-Stream-GenericCloud-9-20240212.0.aarch64.qcow2",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "abc456", Purpose: PurposeShell},
				{Path: "/etc/centos-release", Checksum: "def789", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://cloud.centos.org/centos/9-stream/x86_64/images/CentOS-Stream-GenericCloud-9-20240212.0.x86_64.qcow2",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "abc456", Purpose: PurposeShell},
				{Path: "/etc/centos-release", Checksum: "def789", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   FEDORA
	// ────────────────────────
	FEDORA: {
		"arm64": {
			URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/aarch64/images/Fedora-Cloud-Base-40-1.15.aarch64.raw.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "ghi012", Purpose: PurposeShell},
				{Path: "/etc/fedora-release", Checksum: "jkl345", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://download.fedoraproject.org/pub/fedora/linux/releases/40/Cloud/x86_64/images/Fedora-Cloud-Base-40-1.15.x86_64.raw.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "ghi012", Purpose: PurposeShell},
				{Path: "/etc/fedora-release", Checksum: "jkl345", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   KALI
	// ────────────────────────
	KALI: {
		"arm64": {
			URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-arm64-rootfs.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "mno678", Purpose: PurposeShell},
				{Path: "/etc/kali-version", Checksum: "pqr901", Purpose: PurposeSystemMetadata},
			},
		},
		"amd64": {
			URL: "https://cdimage.kali.org/kali-2024.2/kali-linux-2024.2-amd64-rootfs.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "mno678", Purpose: PurposeShell},
				{Path: "/etc/kali-version", Checksum: "pqr901", Purpose: PurposeSystemMetadata},
			},
		},
	},

	// ────────────────────────
	//   VOID
	// ────────────────────────
	VOID: {
		"arm64": {
			URL: "https://repo-default.voidlinux.org/live/current/void-aarch64-ROOTFS-20240315.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "stu234", Purpose: PurposeShell},
				{Path: "/etc/void-release", Checksum: "vwx567", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://repo-default.voidlinux.org/live/current/void-x86_64-ROOTFS-20240315.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "stu234", Purpose: PurposeShell},
				{Path: "/etc/void-release", Checksum: "vwx567", Purpose: PurposeReleaseInfo},
			},
		},
	},

	// ────────────────────────
	//   OPENSUSE
	// ────────────────────────
	OPENSUSE: {
		"arm64": {
			URL: "https://download.opensuse.org/ports/aarch64/tumbleweed/appliances/openSUSE-Tumbleweed-aarch64-RootFS.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "yz8901", Purpose: PurposeShell},
				{Path: "/etc/SuSE-release", Checksum: "abc234", Purpose: PurposeReleaseInfo},
			},
		},
		"amd64": {
			URL: "https://download.opensuse.org/tumbleweed/appliances/openSUSE-Tumbleweed-x86_64-RootFS.tar.xz",
			Shebangs: []RootFSEntry{
				{Path: "/bin/bash", Checksum: "yz8901", Purpose: PurposeShell},
				{Path: "/etc/SuSE-release", Checksum: "abc234", Purpose: PurposeReleaseInfo},
			},
		},
	},
}

//
// ──────────────────────────────────────────────────────────────
//   HELPERS
// ──────────────────────────────────────────────────────────────
//

func GetRootFS(distro DistroType, arch string) (RootFSConfig, bool) {
	if arch == "" {
		arch = runtime.GOARCH
	}
	m, ok := RootfsCatalog[distro]
	if !ok {
		return RootFSConfig{}, false
	}
	cfg, ok := m[arch]
	return cfg, ok
}

func normalizeArch(arch string) string {
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

func HostArch() string {
	return normalizeArch(runtime.GOARCH)
}

func DetectHostDistros() []DistroType {
	arch := HostArch()
	var supported []DistroType

	for distro, archMap := range RootfsCatalog {
		if _, ok := archMap[arch]; ok {
			supported = append(supported, distro)
		}
	}

	return supported
}

func DetectHostDistroConfig(distro DistroType) (RootFSConfig, error) {
	arch := HostArch()

	archMap, ok := RootfsCatalog[distro]
	if !ok {
		return RootFSConfig{}, fmt.Errorf("❌ distro no soportada: %s", distro)
	}

	cfg, ok := archMap[arch]
	if !ok {
		return RootFSConfig{}, fmt.Errorf(
			"❌ distro %s no soporta arquitectura %s. Soporta: %v",
			distro, arch, availableArchs(distro),
		)
	}

	return cfg, nil
}

func ResolveRootFS(distro DistroType, arch string) (RootFSConfig, error) {
	normArch := normalizeArch(arch)

	cfg, ok := GetRootFS(distro, normArch)
	if !ok {
		// ¿la distro existe?
		_, distroExists := RootfsCatalog[distro]
		if !distroExists {
			return RootFSConfig{}, fmt.Errorf("❌ distro no soportada: %s", distro)
		}

		// El problema es la arquitectura
		return RootFSConfig{}, fmt.Errorf(
			"❌ arquitectura '%s' no soportada para la distro %s. Usa: %v",
			normArch,
			distro,
			availableArchs(distro),
		)
	}

	return cfg, nil
}

// Helper: devuelve las arquitecturas disponibles para una distro
func availableArchs(distro DistroType) []string {
	archs := []string{}
	m := RootfsCatalog[distro]
	for k := range m {
		archs = append(archs, k)
	}
	return archs
}
