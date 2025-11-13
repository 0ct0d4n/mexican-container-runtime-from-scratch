# Axolotl Runtime — Mini Container Engine (Client + SSH Daemon + Ubuntu VM)

Axolotl Runtime is an educational container engine inspired by the early architecture of Docker.  
It is composed of:

- a **client** (running on macOS),  
- a **remote server daemon** (running inside an Ubuntu VM),  
- and the **Linux kernel** features that actually create containers  
  (namespaces, cgroups v2, chroot/pivot_root, overlayfs, and capabilities).

This project is developed on **macOS**, but requires **Ubuntu Linux** as the execution environment — exactly like Docker 0.3–0.7 did, where all containers ran *inside* a separate VM.

---

## 🐧 Why Ubuntu is Required

macOS **cannot** create real Linux containers because it does not use a Linux kernel.  
Containers depend on kernel features such as:

- CLONE_NEWNS, CLONE_NEWPID, CLONE_NEWNET  
- cgroups v2 controllers  
- mount namespaces  
- pivot_root  
- capabilities  
- overlayfs  

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
│  ├─ sets up namespaces       │
│  ├─ overlayfs rootfs (soon)  │
│  └─ executes the container   │
│                              │
│   Linux Kernel = real engine │
└──────────────────────────────┘
```

### Summary:
- **The client** = controller (sends instructions)  
- **The daemon** = executor (creates container resources). The deployment to the VM is automated by invoking the makefile. 
- **Ubuntu + kernel** = actual container engine  

---

## 🧑‍💻 What the Client Does

The client (`axocli`) is the **command issuer**.

It:

- connects to the Ubuntu VM via SSH  
- sends the custom command `AXO_RUN`  
- streams a JSON payload describing:
  - container name  
  - org  
  - cgroup path  
  - memory limit  
  - CPU shares  
  - maximum PIDs  
- waits for the daemon’s response  
- prints success/error information

**In simple terms:**  
🟢 *The client tells the daemon what to create.*  
🔵 *The daemon talks to the Linux kernel to make it real.*  

---

## 📂 Project Structure

```
axolotl/
│
├── daemon/
│   ├── main.go          # SSH daemon
│   ├── cgroup.go        # Cgroups v2 setup
│   ├── filesystem.go    # Overlayfs (coming)
│   ├── namespace.go     # Linux namespaces (coming)
│
├── client/
│   ├── main.go          # The CLI controller
│
├── system/
│   └── axod.service     # Systemd service template
│
├── Makefile             # Build + deploy automation
└── README.md
```

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

## 🚀 Deploying New Changes (mac → ubuntu)

Whenever you update the code:

```bash
export SSHPASS='axolotl'
make deploy
```

The Makefile will:

1. Build axod  
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

./client
```

### Using CLI flags:

```bash
./client   -host 192.168.64.2   -user axolotl   -password axolotl   -insecure   -memory 200   -cpu 0.8
```

---

## 🧪 Verifying Cgroups

```bash
systemd-cgls
```

Or:

```bash
cat /sys/fs/cgroup/axolotl/demo/test-container/memory.max
cat /sys/fs/cgroup/axolotl/demo/test-container/cpu.max
cat /sys/fs/cgroup/axolotl/demo/test-container/pids.max
```

---

## 🧱 Roadmap

### Phase 1 — DONE ✔
- cgroups v2  
- SSH daemon  
- systemd service  
- capability sandboxing  

### Phase 2 — IN PROGRESS 🔧
- Alpine rootfs  
- overlayfs  
- mount namespace + chroot/pivot_root  

### Phase 3 — UPCOMING 🚀
- Network namespaces  
- veth pairs  
- Linux bridge  
- NAT (iptables/nftables)

### Phase 4 — FUTURE 🧬
- `axod run /bin/sh`  
- container lifecycle  
- logs, ps, delete  
- multiple containers support  

---

## ❤️ Long-Term Goal

Eventually this project will have an **installer**, but it will *always* require an external Linux VM — exactly like early Docker — because **only the Linux kernel can run real containers**.

Axolotl Runtime teaches container internals:  
pure syscalls, pure Linux, no shortcuts.

