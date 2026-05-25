# forker

A minimal Linux container runtime written in Go that runs processes inside isolated sandboxes using namespaces and cgroups. This project implements low-level systems programming concepts to construct container isolation layers from scratch without relying on Docker or containerd.

## High-Level Overview

At a high level, `forker` acts as a micro-container engine. When you execute a command, it performs the following sequence:

1. **re-exec trick**: The `forker` parent binary launches itself in a new set of Linux namespaces (UTS, PID, Mount, IPC, Network) via the `syscall.SysProcAttr` configuration.
2. **child initialization**: The re-executed child process configures the hostname, sets up private mounts (binding isolated `/proc` and `/tmp`), and configures networking.
3. **command substitution**: The child process calls `syscall.Exec` to replace itself with the target binary, executing it within the isolated namespaces.
4. **management layer**: The parent process tracks container execution, wires up virtual ethernet links to a host bridge interface, configures resource controls, and manages stops/cleanups.

## Key Implementation Highlights

Here is the low-level work implemented in this runtime:

### 1. Cgroup Unified Hierarchy Integration
- **subtree control delegation**: dynamically writes `+cpu +memory +pids` to `/sys/fs/cgroup/cgroup.subtree_control` and `/sys/fs/cgroup/forker/cgroup.subtree_control` to delegate controllers down the directory tree.
- **resource limitation**: creates individual control groups in `/sys/fs/cgroup/forker/[sandbox_id]/` and applies limits via:
  - `memory.max` for memory limits (e.g., `128M`)
  - `cpu.max` for quota and period limits (e.g., converting cpu float to period/quota shares)
  - `pids.max` to limit process forks and prevent fork bombs
- **process tracking**: registers the spawned container PID in `cgroup.procs` to ensure all child processes are bound by these restrictions.

### 2. Fault-Tolerant Resource Teardown & Zombie Mount Prevention
- **lazy unmounting on panic**: registers a panic-recovery and error handler in `ChildMain` using a deferred execution that calls `syscall.Unmount("/proc", syscall.MNT_DETACH)` and `syscall.Unmount("/tmp", syscall.MNT_DETACH)`. This ensures that even if container setup fails or the binary fails to execute, `/proc` and `/tmp` do not leak as zombie mounts on the host.
- **host cleanup routines**: a defer block tracks container startup phases. If the sandbox fails to become ready within the timeout, the parent process automatically kills the child process, deletes the allocated cgroup, tears down the virtual ethernet interfaces (`veth`), and deletes the sandbox directory.
- **graceful stopping**: `stopSandbox` kills the container process tree using PGID (`-pid`), tears down virtual network links, removes the cgroup path, and deletes runtime state.

### 3. Unix Syscall Type Assertion & Error Inspections
- **precise OS error code extraction**: implements an error parser that type-asserts error objects to `syscall.Errno` (including unwrapping `os.PathError` and `os.SyscallError` structures).
- **logging**: logs precise kernel-level error strings alongside their numeric values (such as `EPERM` for permission errors, `ENOENT` for missing directories, and `EBUSY` for active mounts).

### 4. Cgo-based Pre-Runtime Namespace Joining (`nsenter`)
- **constructor injection**: uses `__attribute__((constructor))` in Cgo to intercept execution before the Go runtime initializes its multi-threaded scheduler.
- **namespace joining**: opens namespace file descriptors under `/proc/[pid]/ns/` and performs `setns` syscalls, allowing additional processes to run inside the exact namespaces of a running sandbox.

---

## Architecture & Namespaces

`forker` creates a separate namespace sandbox using the following Linux namespaces:

| Namespace | syscall flag | Purpose |
|-----------|--------------|---------|
| **UTS** | `CLONE_NEWUTS` | Custom isolated hostname inside sandbox |
| **PID** | `CLONE_NEWPID` | Isolated process tree; container process runs as PID 1 |
| **Mount** | `CLONE_NEWNS` | Isolated mount table; private `/proc` and `/tmp` |
| **IPC** | `CLONE_NEWIPC` | Isolated system V IPC and POSIX message queues |
| **Network** | `CLONE_NEWNET` | Isolated network stack; private `lo` and `eth0` linked via `veth` bridge |

---

## Build and Usage

### Prerequisites
- Linux Kernel (with cgroups enabled)
- Go 1.21+
- Root privileges (necessary to perform `mount`, `unmount`, `setns`, and configure namespaces)
- `iproute2` installed on host

### Building the Project
```bash
go build -o forker ./cmd/forker
```

### Running a Command in a Sandbox
```bash
sudo ./forker run --memory 128M --cpu 0.5 --pids 50 -- /bin/sh
```

### Managing Sandboxes
```bash
# List running sandboxes
sudo ./forker ps

# Execute an additional process inside a sandbox
sudo ./forker exec <sandbox-id> <command> [args...]

# Stop and clean up a sandbox
sudo ./forker stop <sandbox-id>
```

---

## Networking Topology
- **bridge link**: creates a bridge interface named `forker0` with subnet `10.200.0.1/16` on the host.
- **virtual ethernet pair**: allocates a `veth` pair (`veth-[id]` and `veth-ns-[id]`) for each sandbox, moving the namespace peer into the container where it is renamed to `eth0`.
- **ip routing & forwarding**: configures NAT iptables rules and sets `/proc/sys/net/ipv4/ip_forward` on the host to route outbound internet traffic from the container.
