package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const (
	installDir      = ".gopack"
	packageFilename = "go.pack"
	lockFilename    = "go.pack.lock"
)

func printHelpText() {
	fmt.Fprintf(os.Stdout, `Usage: gopack [command] [flags]

Available commands:
	install			Install dependencies
	exec			Execute installed binary
`)
}

func main() {
	if err := run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	args = args[1:] // shift args

	if len(args) == 0 {
		printHelpText()
		return nil
	}

	command := args[0]
	args = args[1:]

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	goPath := filepath.Join(cwd, installDir)
	os.Setenv("GOPATH", goPath)

	switch command {
	case "install":
		return executeInstallCommand(ctx, goPath)
	case "exec":
		if len(args) == 0 {
			return fmt.Errorf("missing binary arg")
		}

		binary := args[0]
		args = args[1:]
		return executeBinaryCommand(ctx, goPath, binary, args)
	default:
		printHelpText()
		return nil
	}
}
