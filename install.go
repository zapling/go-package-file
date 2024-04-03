package main

import (
	"context"
	"debug/buildinfo"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func executeInstallCommand(ctx context.Context, goPath string) error {
	_, err := os.Stat(lockFilename)
	if err == nil {
		return installFromLockfile(ctx, goPath)
	}
	return installFromPackageFile(ctx, goPath)
}

func installFromLockfile(ctx context.Context, goPath string) error {
	f, err := os.Open(lockFilename)
	if err != nil {
		return fmt.Errorf("failed to open lockfile: %w", err)
	}
	defer f.Close()

	entries, err := parseLockfile(f)
	if err != nil {
		return fmt.Errorf("failed to parse lockfile: %w", err)
	}

	var entriesToInstall []lockFileEntry
	_, err = os.Stat(goPath)
	if err != nil {
		entriesToInstall = append(entriesToInstall, entries...)
	} else {
		for _, entry := range entries {
			binaryName := getBinaryName(entry.url)
			filePath := fmt.Sprintf("%s/bin/%s", goPath, binaryName)
			info, err := buildinfo.ReadFile(filePath)
			if err != nil {
				return err
			}

			versionOK := entry.version == info.Main.Version
			sumOK := entry.sum == info.Main.Sum

			if versionOK && sumOK {
				continue
			}

			entriesToInstall = append(entriesToInstall, entry)
		}
	}

	if len(entriesToInstall) == 0 {
		fmt.Fprintf(os.Stdout, "Already up to date\n")
		return nil
	}

	for _, entry := range entriesToInstall {
		urlWithVersion := fmt.Sprintf("%s@%s", entry.url, entry.version)

		cmd := exec.CommandContext(ctx, "go", "install", urlWithVersion)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Fprintf(os.Stdout, "Installing %s\n", urlWithVersion)

		if err := cmd.Run(); err != nil {
			return err
		}

		binaryName := getBinaryName(urlWithVersion)
		info, _ := buildinfo.ReadFile(fmt.Sprintf("%s/bin/%s", goPath, binaryName))
		if entry.sum != info.Main.Sum {
			fmt.Fprintf(
				os.Stderr,
				"Installed binary %s does not match expected sum\n\t%s != %s\n",
				binaryName,
				entry.sum,
				info.Main.Sum,
			)
			return fmt.Errorf("oh no")
		}
	}

	return nil
}

func installFromPackageFile(ctx context.Context, goPath string) error {
	f, err := os.Open(packageFilename)
	if err != nil {
		return fmt.Errorf("failed to open package file: %w", err)
	}
	defer f.Close()

	deps, err := parsePackageFile(f)
	if err != nil {
		return fmt.Errorf("failed to parse package file: %w", err)
	}

	fmt.Fprintf(os.Stdout, "Installing using %s\n", packageFilename)

	for _, dep := range deps {
		cmd := exec.CommandContext(ctx, "go", "install", dep)
		cmd.Stdout = os.Stdout
		cmd.Stdin = os.Stdin

		fmt.Fprintf(os.Stdout, "Installing %s\n", dep)

		if err := cmd.Run(); err != nil {
			return err
		}

	}

	lockF, err := os.OpenFile(lockFilename, os.O_RDWR|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("failed to open lockfile: %w", err)
	}
	defer lockF.Close()

	for _, dep := range deps {
		binaryName := getBinaryName(dep)
		filePath := fmt.Sprintf("%s/bin/%s", goPath, binaryName)
		info, err := buildinfo.ReadFile(filePath)
		if err != nil {
			return err
		}

		_, err = lockF.WriteString(fmt.Sprintf("%s %s %s\n", info.Path, info.Main.Version, info.Main.Sum))
		if err != nil {
			return err
		}
	}

	return nil
}

type lockFileEntry struct {
	url     string
	version string
	sum     string
}

func parseLockfile(r io.Reader) ([]lockFileEntry, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var entries []lockFileEntry
	for _, row := range strings.Split(string(b), "\n") {
		if row == "" {
			continue
		}

		parts := strings.Split(row, " ")
		entries = append(entries, lockFileEntry{
			url:     parts[0],
			version: parts[1],
			sum:     parts[2],
		})
	}

	return entries, nil
}

func parsePackageFile(r io.Reader) ([]string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	var deps []string
	for _, dep := range strings.Split(string(b), "\n") {
		if dep == "" {
			continue
		}
		deps = append(deps, dep)
	}

	return deps, nil
}

func getBinaryName(dependency string) string {
	parts := strings.Split(strings.Split(dependency, "@")[0], "/")
	return parts[len(parts)-1]
}
