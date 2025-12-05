package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gwakko/go-ws-proxy/configs"
	"github.com/gwakko/go-ws-proxy/internal/api"
	"github.com/gwakko/go-ws-proxy/internal/proxy"
	"github.com/gwakko/go-ws-proxy/internal/sse"
	wsHandler "github.com/gwakko/go-ws-proxy/internal/websocket"
)

func main() {
	cfg := configs.Load()
	executor := proxy.NewExecutor(cfg.CommandTimeout)

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
	log.Printf("Server starting on %s", addr)
	log.Printf("  REST:      POST /api/exec")
	log.Printf("  WebSocket: ws://localhost%s/ws", addr)
	log.Printf("  SSE:       GET  /sse/exec?command=...&args=...")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
