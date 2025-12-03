package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"

	"github.com/gwakko/go-ws-proxy/internal/proxy"
)

type Handler struct {
	executor *proxy.Executor
	origins  string
}

func NewHandler(executor *proxy.Executor, origins string) *Handler {
	return &Handler{executor: executor, origins: origins}
}

type WSRequest struct {
	Action  string   `json:"action"`  // "exec"
	Command string   `json:"command"`
	Args    []string `json:"args"`
}

type WSResponse struct {
	Type    string      `json:"type"` // "result", "error", "stream"
	Payload interface{} `json:"payload"`
}

// HandleWS upgrades to WebSocket and handles bidirectional command execution.
func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
	opts := &websocket.AcceptOptions{}
	if h.origins == "*" {
		opts.InsecureSkipVerify = true
	}

	conn, err := websocket.Accept(w, r, opts)
	if err != nil {
		log.Printf("websocket accept error: %v", err)
		return
	}
	defer conn.CloseNow()

	ctx := r.Context()

	for {
		var req WSRequest
		if err := wsjson.Read(ctx, conn, &req); err != nil {
			if websocket.CloseStatus(err) == websocket.StatusNormalClosure {
				return
			}
			log.Printf("ws read error: %v", err)
			return
		}

		switch req.Action {
		case "exec":
			h.handleExec(ctx, conn, req)
		case "stream":
			h.handleStream(ctx, conn, req)
		default:
			wsjson.Write(ctx, conn, WSResponse{
				Type:    "error",
				Payload: "unknown action: " + req.Action,
			})
		}
	}
}

func (h *Handler) handleExec(ctx context.Context, conn *websocket.Conn, req WSRequest) {
	result, err := h.executor.Run(ctx, req.Command, req.Args...)
	if err != nil {
		wsjson.Write(ctx, conn, WSResponse{Type: "error", Payload: err.Error()})
		return
	}
	wsjson.Write(ctx, conn, WSResponse{Type: "result", Payload: result})
}

func (h *Handler) handleStream(ctx context.Context, conn *websocket.Conn, req WSRequest) {
	lines, errs := h.executor.RunStream(ctx, req.Command, req.Args...)

	for line := range lines {
		data, _ := json.Marshal(WSResponse{Type: "stream", Payload: line})
		conn.Write(ctx, websocket.MessageText, data)
	}

	if err := <-errs; err != nil {
		wsjson.Write(ctx, conn, WSResponse{Type: "error", Payload: err.Error()})
	}

	wsjson.Write(ctx, conn, WSResponse{Type: "stream_end", Payload: nil})
}
