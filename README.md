# Axolotl Runtime — Mini Container Engine (Client + SSH Daemon + Ubuntu VM)

Axolotl Runtime is an educational container engine inspired by the early architecture of Docker and built following the patterns of modern container runtimes like **runc** and **containerd**.

It is composed of:

- a **client** (running on macOS),
- a **remote server daemon** (running inside an Ubuntu VM),
- and the **Linux kernel** features that actually create containers
  (namespaces, cgroups v2, pivot_root, veth pairs, and capabilities).

This project is developed on **macOS**, but requires **Ubuntu Linux** as the execution environment — exactly like Docker 0.3–0.7 did, where all containers ran *inside* a separate VM.

---

## 🐧 Why Ubuntu is Required

macOS **cannot** create real Linux containers because it does not use a Linux kernel.
Containers depend on kernel features such as:

- CLONE_NEWNS, CLONE_NEWPID, CLONE_NEWNET, CLONE_NEWUTS, CLONE_NEWIPC
- cgroups v2 controllers
- mount namespaces
- pivot_root
- veth pairs & network namespaces
- capabilities

These features exist only in **the Linux kernel**, so the daemon must run inside:

### ✔ A dedicated Ubuntu VM
Any of the following works:

- UTM
- VirtualBox
- Parallels
- VMWare Fusion
- Lima/Colima
- Docker Desktop's internal Linux VM

Your mac is the **development machine**.
Ubuntu is where the runtime **actually runs**.

---

## 🧱 Project Architecture

```
┌──────────────────────────────┐
│          macOS Host           │
│    (Developer Environment)    │
│                              │
│  axocli (client)             │
│   └─ sends AXO_RUN → SSH →   │
└──────────────┬───────────────┘
               │
               ▼
┌──────────────────────────────┐
│         Ubuntu VM             │
│       (Required Linux)        │
│                              │
│  axod (daemon)               │
│  ├─ listens on port 2222     │
│  ├─ creates cgroups v2       │
│  ├─ applies CPU/memory/pids  │
│  ├─ downloads & extracts rootfs
│  ├─ sets up namespaces       │
│  ├─ pivot_root filesystem    │
│  ├─ creates veth pairs       │
│  └─ executes the container   │
│                              │
│   Linux Kernel = real engine │
└──────────────────────────────┘
```

### Summary:
- **The client** = controller (sends instructions)
- **The daemon** = executor (creates container resources). The deployment to the VM is automated by invoking the makefile.
- **libaxolotl** = reusable container runtime library (like libcontainer in runc)
- **Ubuntu + kernel** = actual container engine

---

## 📂 Project Structure (v2.0 - Refactored)

Following modern Go best practices and the patterns used by **runc** and **containerd**:

```
axolotl/
│
├── cmd/                           # Entry points
│   ├── axod/                      # Daemon entry point
│   │   ├── main.go                # SSH server + init-container
│   │   ├── debug_on.go
│   │   └── debug_off.go
│   └── axocli/                    # Client CLI entry point
│       └── main.go
│
├── libaxolotl/                    # 📚 Public runtime library (like libcontainer)
│   ├── container.go               # Container orchestrator
│   ├── init_linux.go              # Init process (PID 1)
│   │
│   ├── types/                     # Public types & specs
│   │   └── types.go               # Container specs, namespace configs
│   │
│   ├── cgroups/                   # Cgroups v2 management
│   │   ├── manager.go             # Cgroup lifecycle
│   │   └── resources.go           # CPU, memory, PIDs limits
│   │
│   ├── network/                   # Networking
│   │   ├── veth.go                # veth pair creation
│   │   ├── netns.go               # Network namespace operations
│   │   └── bridge.go              # Bridge management
│   │
│   └── rootfs/                    # Root filesystem management
│       ├── distro.go              # Distro catalog (Alpine, Ubuntu, etc.)
│       ├── downloader.go          # Download rootfs tarballs
│       ├── manager.go             # Rootfs orchestration
│       ├── mount.go               # Mount proc, sys, dev, tmp
│       └── pivot.go               # pivot_root implementation
│
├── internal/                      # 🔒 Private implementation
│   ├── daemon/                    # SSH daemon logic
│   │   └── handler.go             # Connection & command handling
│   ├── client/                    # Client implementation
│   │   ├── client.go
│   │   ├── config.go
│   │   ├── builder.go
│   │   └── response.go
│   ├── command/                   # Command definitions
│   └── util/                      # Utilities
│       ├── util.go                # Tar extraction, ID generation
│       ├── rootfs/
│       └── tini/
│
├── Makefile                       # Build + deploy automation
└── README.md
```

### 🎯 Architecture Principles

- **`cmd/`** - Entry points for binaries (axod, axocli)
- **`libaxolotl/`** - Public library that can be imported by other projects
- **`internal/`** - Private code (Go enforces it cannot be imported externally)

This structure follows the same patterns as:
- `runc` → `libcontainer/`
- `containerd` → `pkg/`

---

## 🧑‍💻 What the Client Does

The client (`axocli`) is the **command issuer**.

It:

- connects to the Ubuntu VM via SSH
- sends the custom command `AXO_RUN`
- streams a JSON payload describing:
  - container name
  - org
  - image (alpine, ubuntu, etc.)
  - cgroup path
  - memory limit
  - CPU shares
  - maximum PIDs
  - entrypoint command
- waits for the daemon's response
- prints success/error information

**In simple terms:**
🟢 *The client tells the daemon what to create.*
🔵 *The daemon talks to the Linux kernel to make it real.*

---

## 🔧 Requirements (macOS + Ubuntu)

### On macOS (development):

```bash
brew install go sshpass
```

Required because the Makefile uses `sshpass` to automate deploys.

### On Ubuntu VM:

- Ubuntu 20.04+
- cgroup v2 fully enabled
- SSH server
- systemd
- Kernel ≥ 5.0

---

## 🏗️ Building the Project

### Build daemon for Linux ARM64:

```bash
make build
# or manually:
GOOS=linux GOARCH=arm64 go build -o axod ./cmd/axod
```

### Build client for macOS:

```bash
go build -o axocli ./cmd/axocli
```

---

## 🚀 Deploying New Changes (mac → ubuntu)

Whenever you update the code:

```bash
export SSHPASS='axolotl'
make deploy
```

The Makefile will:

1. Build axod for Linux ARM64
2. Upload using SCP
3. Install into `/usr/local/bin`
4. Reload systemd
5. Restart the service
6. Show logs

---

## 📌 Example Client Usage

### Using environment variables:

```bash
export AXOLOTL_HOST=192.168.64.2
export AXOLOTL_PORT=2222
export AXOLOTL_USER=axolotl
export AXOLOTL_PASSWORD=axolotl
export AXOLOTL_INSECURE_HOST_KEY=true
export AXOLOTL_TIMEOUT=30s

./axocli
```

### Using CLI flags:

```bash
./axocli -host 192.168.64.2 -user axolotl -password axolotl -insecure -memory 200 -cpu 0.8 -image alpine
```

---

## 🧪 Verifying Container Resources

### Check cgroups:

```bash
systemd-cgls
```

Or directly:

```bash
cat /sys/fs/cgroup/global_test2_test/memory.max
cat /sys/fs/cgroup/global_test2_test/cpu.max
cat /sys/fs/cgroup/global_test2_test/pids.max
```

### Check network namespaces:

```bash
# Inside container
ip addr show
ip link show

# On host
ip netns list
ip link show | grep veth
```

---

## 🧱 Roadmap

### Phase 1 — DONE ✔
- ✅ cgroups v2 (CPU, memory, PIDs)
- ✅ SSH daemon with custom commands
- ✅ systemd service integration
- ✅ Linux namespaces (PID, Mount, UTS, IPC, Network)
- ✅ rootfs download & extraction (Alpine, Ubuntu, Debian)
- ✅ pivot_root implementation
- ✅ veth pair creation
- ✅ Init process (PID 1) with zombie reaping
- ✅ Signal forwarding (tini-like)
- ✅ Refactored architecture (libaxolotl pattern)

### Phase 2 — IN PROGRESS 🔧
- 🔧 Linux bridge setup
- 🔧 Container-to-host networking
- 🔧 Multiple containers support
- 🔧 OverlayFS for efficient storage

### Phase 3 — UPCOMING 🚀
- NAT & port forwarding
- Container-to-container networking
- DNS resolution
- Volume mounts
- Environment variables

### Phase 4 — FUTURE 🧬
- Container lifecycle management (start, stop, restart)
- Logs streaming
- Container inspection
- Health checks
- Resource monitoring

---

## 🏗️ Container Lifecycle (Current Implementation)

```
1. Client sends AXO_RUN command
   ↓
2. Daemon creates cgroup & applies limits
   ↓
3. Daemon downloads & extracts rootfs
   ↓
4. Daemon spawns child process with namespaces
   (CLONE_NEWNS | CLONE_NEWPID | CLONE_NEWUTS | CLONE_NEWIPC | CLONE_NEWNET)
   ↓
5. Host configures networking (veth pair)
   ↓
6. Init process (PID 1) inside container:
   - Mounts proc, sys, dev, tmp
   - Performs pivot_root
   - Configures container-side networking
   - Starts zombie reaper
   - Executes user command
   ↓
7. Wait for container exit
```

---

## 📚 Learning Resources

This project is designed to teach container internals. Key concepts:

- **Namespaces** - Process isolation
- **Cgroups v2** - Resource limits
- **pivot_root** - Secure filesystem isolation (better than chroot)
- **veth pairs** - Virtual network interfaces
- **PID 1** - Init process responsibilities
- **Zombie reaping** - Cleaning up orphaned processes

---

## 🔄 Version 2.0 - Breaking Changes

This version includes a **major refactoring** following modern Go practices:

### What Changed:
- ❌ Removed `pkg/` directory
- ✅ Added `cmd/` for entry points
- ✅ Added `libaxolotl/` as public runtime library
- ✅ Added `internal/` for private code
- ✅ Better separation of concerns
- ✅ Cleaner imports and module structure

### Migration Guide:
If you were importing this project:
```go
// Old
import "axolotl/pkg/model"
import "axolotl/pkg/server"

// New
import "axolotl/libaxolotl/types"
import "axolotl/libaxolotl"
```

---

## ❤️ Long-Term Goal

Eventually this project will have an **installer**, but it will *always* require an external Linux VM — exactly like early Docker — because **only the Linux kernel can run real containers**.

Axolotl Runtime teaches container internals:
pure syscalls, pure Linux, no shortcuts.

**Inspired by:**
- Docker 0.x architecture
- runc's libcontainer
- containerd's modular design

---

## 📄 License

Educational project - MIT License
