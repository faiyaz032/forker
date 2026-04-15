package runtime

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

const childEnv = "__FORKER_CHILD__"

func IsChildProcess() bool {
	return os.Getenv(childEnv) == "1"
}

func Run(args []string) error {
	if len(args) < 2 {
		Usage()
		return fmt.Errorf("invalid args")
	}

	switch args[1] {
	case "run":
		if len(args) < 3 {
			return fmt.Errorf("missing command")
		}
		command := args[2]
		commandArgs := args[3:]
		return runInNamespace(command, commandArgs)

	case "ps":
		return listSandboxes()

	default:
		Usage()
		return fmt.Errorf("unknown command")
	}
}
func Usage() {
	fmt.Println(`forker run <command> [args...]`)
}

func runInNamespace(command string, args []string) error {
	bin, err := exec.LookPath(command)
	if err != nil {
		return fmt.Errorf("cannot find %q: %w", command, err)
	}

	bin, _ = filepath.Abs(bin)

	sandboxID := fmt.Sprintf("forker-%04x", rand.Intn(0xffff))

	self, err := os.Executable()
	if err != nil {
		return err
	}

	fmt.Printf("[forker] starting %q in sandbox %s\n", bin, sandboxID)

	childCmd := exec.Command(self)

	childCmd.Stdin = os.Stdin

	childCmd.Stdout = nil
	childCmd.Stderr = nil

	childCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,

		Setsid: true,
	}

	argsJson, err := json.Marshal(args)
	if err != nil {
		return err
	}

	childCmd.Env = append(os.Environ(),
		childEnv+"=1",
		"__FORKER_BIN__="+bin,
		"__FORKER_ARGS__="+string(argsJson),
		"__FORKER_HOSTNAME__="+sandboxID,
		"__FORKER_SANDBOX_ID__="+sandboxID,
	)

	if err := childCmd.Start(); err != nil {
		return err
	}

	fmt.Printf("[forker] sandbox %s started (pid=%d)\n", sandboxID, childCmd.Process.Pid)

	return nil

}

func saveSandbox(id string, pid int, cmd string) error {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return err
	}

	dir := filepath.Join(basePath, id)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(
		filepath.Join(dir, "pid"),
		[]byte(fmt.Sprintf("%d", pid)),
		0644,
	)
}
