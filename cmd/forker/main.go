package main

import (
	"fmt"
	"os"

	"github.com/faiyaz032/forker/internal/runtime"
)

func main() {
	//check if this a child process
	if runtime.IsChildProcess() {
		if err := runtime.ChildMain(); err != nil {
			fmt.Fprintf(os.Stderr, "[forker child] error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// parent process
	if err := runtime.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
