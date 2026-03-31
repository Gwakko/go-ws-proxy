package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gwakko/go-ws-proxy/configs"
	"github.com/gwakko/go-ws-proxy/internal/api"
	"github.com/gwakko/go-ws-proxy/internal/proxy"
	"github.com/gwakko/go-ws-proxy/internal/sse"
	wsHandler "github.com/gwakko/go-ws-proxy/internal/websocket"
)

var defaultAllowedCommands = []string{
	"ping", "curl", "ls", "df", "uptime", "hostname", "whoami", "dig", "nslookup",
}

type allowlistFile struct {
	Commands []string `json:"commands"`
}

func loadAllowlist(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loadAllowlist: reading %s: %w", path, err)
	}

	var af allowlistFile
	if err := json.Unmarshal(data, &af); err != nil {
		return nil, fmt.Errorf("loadAllowlist: parsing %s: %w", path, err)
	}

	if len(af.Commands) == 0 {
		return nil, fmt.Errorf("loadAllowlist: no commands found in %s", path)
	}

	return af.Commands, nil
}

func main() {
	cfg := configs.Load()

	allowed, err := loadAllowlist(cfg.AllowlistPath)
	if err != nil {
		log.Printf("Warning: could not load allowlist from %s: %v; using defaults", cfg.AllowlistPath, err)
		allowed = defaultAllowedCommands
	} else {
		log.Printf("Loaded %d commands from allowlist %s", len(allowed), cfg.AllowlistPath)
	}

	executor := proxy.NewExecutor(cfg.CommandTimeout, allowed)

	apiH := api.NewHandler(executor)
	wsH := wsHandler.NewHandler(executor, cfg.AllowedOrigins)
	sseH := sse.NewHandler(executor)

	mux := http.NewServeMux()

	// REST endpoints
	mux.HandleFunc("/api/health", apiH.HandleHealth)
	mux.HandleFunc("/api/exec", apiH.HandleExec)

	// WebSocket endpoint
	mux.HandleFunc("/ws", wsH.HandleWS)

	// SSE endpoint
	mux.HandleFunc("/sse/exec", sseH.HandleSSE)

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", addr)
		log.Printf("  REST:      POST /api/exec")
		log.Printf("  WebSocket: ws://localhost%s/ws", addr)
		log.Printf("  SSE:       GET  /sse/exec?command=...&args=...")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
