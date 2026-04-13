package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"dagger.io/dagger"
)

// ASRServer manages the Dagger-based Python ASR service lifecycle.
type ASRServer struct {
	client   *dagger.Client
	service  *dagger.Service
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
		WithExec([]string{"sh", "-c", "apt-get update && apt-get install -y --no-install-recommends git ffmpeg && rm -rf /var/lib/apt/lists/*"}).
		WithMountedCache("/root/.cache/huggingface", hfCache).
		WithMountedCache("/root/.cache/pip", pipCache).
		WithDirectory("/app", serverDir).
		WithWorkdir("/app").
		WithExec([]string{"pip", "install", "-r", "requirements.txt"}).
		WithExposedPort(opts.Port).
		WithExec([]string{
			"uvicorn", "server:app",
			"--host", "0.0.0.0",
			"--port", fmt.Sprintf("%d", opts.Port),
		})

	// Convert to a Dagger service and start it
	service := ctr.AsService()

	// Start the service explicitly — this blocks until the exposed port accepts connections,
	// which means model loading in the FastAPI lifespan must complete first.
	log.Printf("Starting container (pip + model load from cache)...")
	service, err = service.Start(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("start service: %w", err)
	}
	log.Printf("Container started, establishing tunnel...")

	// Create a tunnel from the host to the running service
	tunnel := client.Host().Tunnel(service)
	endpoint, err := tunnel.Endpoint(ctx, dagger.ServiceEndpointOpts{Port: opts.Port})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("get tunnel endpoint: %w", err)
	}

	svc := &ASRServer{
		client:   client,
		service:  service,
		endpoint: endpoint,
	}

	// Wait for the HTTP server to be ready (model loading can take time)
	log.Printf("ASR server ready at %s", s.endpoint)
	if err := svc.waitReady(ctx); err != nil {
		svc.Stop()
		return nil, fmt.Errorf("server health check: %w", err)
	}

	return svc, nil
}

func (s *ASRServer) waitReady(ctx context.Context) error {
	url := fmt.Sprintf("http://%s/health", s.endpoint)
	for i := 0; i < 60; i++ { // 60 second timeout
		resp, err := http.Get(url)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				log.Printf("Health check passed (attempt %d)", i+1)
				return nil
			}
			log.Printf("Health check status %d (attempt %d)", resp.StatusCode, i+1)
		} else {
			if i < 5 || i%10 == 0 {
				log.Printf("Health check failed (attempt %d): %v", i+1, err)
			}
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

// Stop tears down the service and closes the Dagger client.
func (s *ASRServer) Stop() {
	if s.service != nil {
		s.service.Stop(context.Background())
	}
	if s.client != nil {
		s.client.Close()
	}
}
