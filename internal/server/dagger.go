package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"dagger.io/dagger"
)

// ASRServer manages the Dagger-based Python ASR service lifecycle.
type ASRServer struct {
	client   *dagger.Client
	endpoint string
}

// Options configures the ASR server.
type Options struct {
	ServerDir string
	Port      int
}

// Start launches the Python ASR server as a Dagger Service, creates a host tunnel,
// and waits for the health check to pass.
func Start(ctx context.Context, opts Options) (*ASRServer, error) {
	if opts.Port == 0 {
		opts.Port = 8000
	}

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return nil, fmt.Errorf("connect dagger: %w", err)
	}

	hfCache := client.CacheVolume("transcription-hf-cache")
	pipCache := client.CacheVolume("transcription-pip-cache")

	serverDir := client.Host().Directory(opts.ServerDir,
		dagger.HostDirectoryOpts{Exclude: []string{"__pycache__", "*.pyc"}})

	ctr := client.Container().
		From("python:3.11-slim-bookworm").
		WithMountedCache("/root/.cache/huggingface", hfCache).
		WithMountedCache("/root/.cache/pip", pipCache).
		WithDirectory("/app", serverDir).
		WithWorkdir("/app").
		WithExec([]string{"pip", "install", "--no-cache-dir", "-r", "requirements.txt"}).
		WithExposedPort(opts.Port).
		WithExec([]string{
			"uvicorn", "server:app",
			"--host", "0.0.0.0",
			"--port", fmt.Sprintf("%d", opts.Port),
		})

	service := ctr.AsService()
	tunnel := client.Host().Tunnel(service)

	endpoint, err := tunnel.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: opts.Port})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("get tunnel endpoint: %w", err)
	}

	svc := &ASRServer{
		client:   client,
		endpoint: endpoint,
	}

	// Wait for server to be ready
	if err := svc.waitReady(ctx); err != nil {
		svc.Stop()
		return nil, fmt.Errorf("server not ready: %w", err)
	}

	return svc, nil
}

func (s *ASRServer) waitReady(ctx context.Context) error {
	for i := 0; i < 120; i++ { // 120 second timeout (pip install can be slow)
		resp, err := http.Get(fmt.Sprintf("http://%s/health", s.endpoint))
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
		}
	}
	return fmt.Errorf("server not ready after 120s")
}

// Endpoint returns the host:port address for the ASR server tunnel.
func (s *ASRServer) Endpoint() string {
	return s.endpoint
}

// Stop closes the Dagger client, which tears down the service and tunnel.
func (s *ASRServer) Stop() {
	if s.client != nil {
		s.client.Close()
	}
}
