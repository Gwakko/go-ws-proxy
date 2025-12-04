package sse

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gwakko/go-ws-proxy/internal/proxy"
)

type Handler struct {
	executor *proxy.Executor
}

func NewHandler(executor *proxy.Executor) *Handler {
	return &Handler{executor: executor}
}

// HandleSSE streams command output as Server-Sent Events.
// GET /sse/exec?command=ping&args=-c,4,8.8.8.8
func (h *Handler) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	command := r.URL.Query().Get("command")
	if command == "" {
		http.Error(w, "command parameter is required", http.StatusBadRequest)
		return
	}

	var args []string
	if argsStr := r.URL.Query().Get("args"); argsStr != "" {
		args = strings.Split(argsStr, ",")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	flusher.Flush()

	lines, errs := h.executor.RunStream(r.Context(), command, args...)

	for line := range lines {
		fmt.Fprintf(w, "data: %s\n\n", strings.TrimRight(line, "\n"))
		flusher.Flush()
	}

	if err := <-errs; err != nil {
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
		flusher.Flush()
		log.Printf("sse stream error: %v", err)
	}

	fmt.Fprintf(w, "event: done\ndata: stream finished\n\n")
	flusher.Flush()
}
