package api

import (
	"encoding/json"
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

type ExecRequest struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

// HandleExec handles POST /api/exec — run a command and return result as JSON.
func (h *Handler) HandleExec(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ExecRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		http.Error(w, "command is required", http.StatusBadRequest)
		return
	}

	result, err := h.executor.Run(r.Context(), req.Command, req.Args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// HandleHealth handles GET /api/health.
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
