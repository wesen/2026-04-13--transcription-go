package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

const DefaultPort = 8000

// ResolveServerDir resolves the Python server directory for both normal binaries
// and `go run` executions. The returned directory must contain server.py.
func ResolveServerDir(serverDir string) (string, error) {
	if filepath.IsAbs(serverDir) {
		return validateServerDir(serverDir)
	}

	if exePath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exePath), "..", "..", serverDir)
		if resolved, err := validateServerDir(candidate); err == nil {
			return resolved, nil
		}
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}
	return validateServerDir(filepath.Join(cwd, serverDir))
}

// StartDefault starts the ASR server with repository defaults suitable for CLI use.
func StartDefault(ctx context.Context, serverDir string) (*ASRServer, error) {
	return Start(ctx, Options{
		ServerDir: serverDir,
		Port:      DefaultPort,
	})
}

func validateServerDir(serverDir string) (string, error) {
	serverPy := filepath.Join(serverDir, "server.py")
	if _, err := os.Stat(serverPy); err != nil {
		return "", fmt.Errorf("server.py not found in %s: %w", serverDir, err)
	}
	return filepath.Clean(serverDir), nil
}
