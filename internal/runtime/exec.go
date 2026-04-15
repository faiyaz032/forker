package runtime

import (
	"encoding/json"
	"os"
	"os/exec"
	"syscall"
)

func execInSandbox(id, command string, args []string) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}

	jsonArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}

	cmd := exec.Command(self, "exec-child")

	cmd.Env = append(os.Environ(),
		"__FORKER_EXEC__=1",
		"__FORKER_ID__="+id,
		"__FORKER_CMD__="+command,
		"__FORKER_ARGS__="+string(jsonArgs),
	)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func ExecChild() error {
	command := os.Getenv("__FORKER_CMD__")

	var args []string
	if err := json.Unmarshal([]byte(os.Getenv("__FORKER_ARGS__")), &args); err != nil {
		return err
	}

	absCmd, err := exec.LookPath(command)
	if err != nil {
		return err
	}

	finalArgs := append([]string{command}, args...)
	return syscall.Exec(absCmd, finalArgs, os.Environ())
}
