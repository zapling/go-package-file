package main

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
)

func executeBinaryCommand(ctx context.Context, goPath string, binary string, args []string) error {
	path := filepath.Join(goPath, "bin")
	fullPath := path + "/./" + binary

	cmd := exec.CommandContext(ctx, fullPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
