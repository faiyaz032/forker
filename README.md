# forker

A minimal Linux container runtime written in Go. It runs programs in isolated sandboxes using namespaces and cgroups. This is built from scratch using Go and raw Linux system calls to show how containers work under the hood.

## High Level Overview

At a high level, `forker` acts as a tiny container engine. When you run a command:

1. **re-exec trick**: The parent process starts a child process in new isolated Linux namespaces.
2. **child init**: The child process sets the custom hostname, mounts private directories, and brings up network interfaces.
3. **substitution**: The child process replaces itself with the target command so it runs fully isolated.
4. **management**: The parent process maps the network to the host, configures resource limits, and cleans up when finished.

## Key Implementation Highlights

Here is the low level work implemented in this runtime:

### 1. Cgroup Resource Limits
- **limits mapping**: Creates control groups under `/sys/fs/cgroup` to set limits for sandbox processes.
- **resource bounds**: Controls maximum memory use, CPU quotas, and the total number of processes to prevent system crashes.
- **process tracking**: Attaches the container process to the cgroup folder so all its children are bounded.

### 2. Failure Cleanup and Mount Teardown
- **mount cleanup**: Uses deferred code inside the sandbox child to unmount the `/proc` and `/tmp` filesystems if a crash or exit occurs. This prevents leaving behind zombie mounts on the host system.
- **host teardown**: If the container fails to start, the parent process automatically kills the child, deletes the cgroup folder, removes the virtual ethernet interfaces, and clears the state.
- **clean stops**: When stopping a sandbox, it removes all temporary virtual interfaces and files completely.

### 3. Syscall Error Checking
- **error verification**: Inspects errors from raw operating system calls and reads the exact error codes like EPERM for permission issues or ENOENT for missing files.
- **logging**: Displays human friendly operating system error codes directly to make debugging easier.

### 4. Joining Existing Sandboxes
- **namespace joining**: Uses C code via Cgo to let new processes enter the exact namespaces of an already running sandbox.
- **re-entry**: Opens namespace files under `/proc` to run additional commands inside the same isolation boundaries.

---

## Namespaces and Isolation

| Namespace | Syscall Flag | Purpose |
|-----------|--------------|---------|
| UTS | CLONE_NEWUTS | Custom hostname inside the sandbox |
| PID | CLONE_NEWPID | Process tree isolation where container process is PID 1 |
| Mount | CLONE_NEWNS | Isolated mount table so directory changes do not leak |
| IPC | CLONE_NEWIPC | Isolated inter-process communications |
| Network | CLONE_NEWNET | Isolated network interfaces linked to the host |

---

## Build and Usage

### Prerequisites
- Linux with cgroups enabled
- Go 1.21 or newer
- Root or sudo privileges
- iproute2 tool installed on host

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
# list running sandboxes
sudo ./forker ps

# run another command inside a sandbox
sudo ./forker exec <sandbox-id> <command> [args...]

# stop and clean up a sandbox
sudo ./forker stop <sandbox-id>
```

---

## Networking Topology
- **bridge interface**: Sets up a virtual network bridge named `forker0` on the host.
- **ethernet pairs**: Allocates a pair of virtual ethernet interfaces to link the host bridge directly into the isolated sandbox network.
- **traffic routing**: Sets up iptables rules to route internet traffic between the host and the sandbox.
