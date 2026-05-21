# forker

A minimal Linux container runtime written in Go that runs processes inside isolated sandboxes using Linux namespaces — no Docker, no containerd, just raw syscalls.

## What It Does

`forker` launches any command inside an isolated environment with its own:

- **Hostname** (UTS namespace)
- **Process tree** — PIDs start at 1 inside the sandbox (PID namespace)
- **Filesystem mounts** — private `/proc` and `/tmp` (Mount namespace)
- **Network stack** — isolated loopback interface (Network namespace)
- **IPC** (IPC namespace)

It uses the **re-exec trick**: the same binary runs twice — once as the parent to set up namespaces, and again as the child to initialize and run inside the sandbox.

## Usage

```bash
# Build
go build -o forker ./cmd/forker

# Run a command in a sandbox
sudo ./forker run <command> [args...]

# Run in detached (background) mode
sudo ./forker run -d <command> [args...]
```

**Example (Go Web Server):**

Create a simple Go file (e.g., `server.go`):
```go
package main
import ("fmt"; "net/http"; "os")
func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		fmt.Fprintf(w, "Hello from sandbox! Hostname: %s, PID: %d\n", hostname, os.Getpid())
	})
	fmt.Println("Server starting on :8080...")
	http.ListenAndServe(":8080", nil)
}
```

Then run it with `forker`:
```bash
sudo ./forker run go run server.go
```

This starts the Go server inside an isolated sandbox. You can then visit `localhost:8080` or `curl localhost:8080` from another terminal on your host.

> **Note:** `sudo` is required because creating Linux namespaces needs elevated privileges.

## How It Works

```
forker run go run server.go
    │
    ▼ (parent process)
Parses args → generates sandboxID (e.g. forker-3a7f)
Re-executes itself with Linux CLONE_NEW* flags
Sets config via environment variables
    │
    ▼ (child process, inside new namespaces)
Sets hostname   → "forker-3a7f"
Sets up mounts  → private /proc, tmpfs /tmp
Sets up network → brings up loopback (lo)
Starts program  → go run server.go
Drops into      → /bin/bash (interactive shell)
```

## Project Structure

```
forker/
├── cmd/
│   └── forker/
│       └── main.go          # Entry point — routes parent vs child execution
└── internal/
    └── runtime/
        ├── run.go           # Parses CLI args, creates namespaces, re-execs self
        ├── child.go         # Child init: loads config, sets up sandbox, starts program
        ├── namespace.go     # Sets hostname via syscall
        ├── mount.go         # Configures /proc and /tmp mounts
        └── network.go       # Brings up loopback interface (ip link set lo up)
```

## Namespaces Used

| Namespace | Flag | Effect |
|-----------|------|--------|
| UTS | `CLONE_NEWUTS` | Custom hostname inside sandbox |
| PID | `CLONE_NEWPID` | Process IDs isolated; container PID 1 |
| Mount | `CLONE_NEWNS` | Filesystem mounts don't leak in/out |
| IPC | `CLONE_NEWIPC` | Isolated inter-process communication |
| Network | `CLONE_NEWNET` | Isolated network stack |

## Requirements

- Linux (namespaces are a Linux kernel feature)
- Go 1.21+
- `root` / `sudo` privileges
- `ip` command available (`iproute2` package)

## Inspiration

This project is inspired by how production runtimes like `runc` and `containerd` work under the hood. It's intentionally minimal for learning purposes.
