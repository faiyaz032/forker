package main

import (
	"fmt"
	"os"

	"github.com/faiyaz032/forker/internal/runtime"
)

func main() {
	// ---------------------------
	// EXEC MODE (NEW)
	// ---------------------------
	if len(os.Args) > 1 && os.Args[1] == "exec-child" {
		if err := runtime.ExecChild(); err != nil {
			fmt.Fprintf(os.Stderr, "[forker exec-child] error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ---------------------------
	// CHILD MODE (sandbox init)
	// ---------------------------
	if runtime.IsChildProcess() {
		if err := runtime.ChildMain(); err != nil {
			fmt.Fprintf(os.Stderr, "[forker child] error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// ---------------------------
	// PARENT MODE (forker daemon)
	// ---------------------------
	if err := runtime.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "[forker] error: %v\n", err)
		os.Exit(1)
	}
}
