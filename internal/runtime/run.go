package runtime

import (
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
	if len(args) < 3 || args[1] != "run" {
		Usage()
		return fmt.Errorf("invalid args")
	}

	detached := false
	i := 2

	if args[i] == "-d" {
		detached = true
		i++
	}

	if i >= len(args) {
		Usage()
		return fmt.Errorf("missing command")
	}

	command := args[i]
	commandArgs := args[i+1:]

	return runInNamespace(command, commandArgs, detached)
}

func Usage() {
	fmt.Println(`forker run [-d] <command> [args...]`)
}

func runInNamespace(command string, args []string, detached bool) error {
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

	if detached {
		childCmd.Stdout = nil
		childCmd.Stderr = nil
	} else {
		childCmd.Stdout = os.Stdout
		childCmd.Stderr = os.Stderr
	}

	childCmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWNS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWNET,

		Setsid: true,
	}

	childCmd.Env = append(os.Environ(),
		childEnv+"=1",
		"__FORKER_BIN__="+bin,
		"__FORKER_ARGS__="+encodeArgs(args),
		"__FORKER_HOSTNAME__="+sandboxID,
		"__FORKER_SANDBOX_ID__="+sandboxID,
	)

	if detached {
		return childCmd.Start()
	}

	return childCmd.Run()
}

func encodeArgs(args []string) string {
	out := ""
	for i, a := range args {
		if i > 0 {
			out += "\x00"
		}
		out += a
	}
	return out
}

func decodeArgs(s string) []string {
	if s == "" {
		return nil
	}

	var res []string
	cur := ""

	for i := 0; i < len(s); i++ {
		if s[i] == 0 {
			res = append(res, cur)
			cur = ""
		} else {
			cur += string(s[i])
		}
	}

	return append(res, cur)
}
