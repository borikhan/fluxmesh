package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/borikhan/fluxmesh/mesh"
)

func main() {
	worker := mesh.Worker{
		Retry:   mesh.RetryPolicy{Attempts: 3, BaseDelay: 25 * time.Millisecond, MaxDelay: time.Second},
		Breaker: &mesh.Breaker{Threshold: 5, Cooldown: 15 * time.Second}, Deduper: mesh.NewDeduper(10 * time.Minute),
		Handler: func(_ context.Context, m mesh.Message) error {
			slog.Info("message processed", "id", m.ID, "topic", m.Topic)
			return nil
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /v1/messages", func(w http.ResponseWriter, r *http.Request) {
		var m mesh.Message
		if json.NewDecoder(r.Body).Decode(&m) != nil || m.ID == "" {
			http.Error(w, "invalid message", 400)
			return
		}
		if err := worker.Process(r.Context(), m); err != nil {
			http.Error(w, err.Error(), 503)
			return
		}
		w.WriteHeader(202)
	})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	slog.Info("worker listening", "port", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		slog.Error("worker stopped", "error", err)
	}
}
