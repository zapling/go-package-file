package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func executeInstallCommand(ctx context.Context) error {
	f, err := os.Open(packageFile)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	dependencies := strings.Split(string(b), "\n")
	for _, dep := range dependencies {
		if dep == "" {
			continue
		}

		fmt.Printf("Installing %s\n", dep)

		installCmd := exec.CommandContext(ctx, "go", "install", dep)
		installCmd.Stdout = os.Stdout
		installCmd.Stderr = os.Stderr

		if err := installCmd.Run(); err != nil {
			return err
		}
	}

	return nil
}
