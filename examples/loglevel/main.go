package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/bigboss2063/sugarzero"
)

const (
	apiAddr      = ":8080"
	defaultLevel = "info"
)

var (
	allowedLevels = []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	allowedLookup = func() map[string]struct{} {
		m := make(map[string]struct{}, len(allowedLevels))
		for _, lvl := range allowedLevels {
			m[lvl] = struct{}{}
		}
		return m
	}()
)

type (
	levelChangeRequest struct {
		Level string `json:"level"`
	}

	levelResponse struct {
		CurrentLevel string   `json:"current_level"`
		DefaultLevel string   `json:"default_level"`
		Allowed      []string `json:"allowed_levels"`
	}

	errorResponse struct {
		Error string `json:"error"`
	}
)

func main() {
	ctx, err := sugarzero.New(context.Background(), defaultLevel)
	if err != nil {
		panic(err)
	}

	go emitBackgroundLogs(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/log-level", logLevelHandler(ctx))

	sugarzero.Infof(ctx, "log level API listening on http://localhost%s/log-level", apiAddr)

	if err := http.ListenAndServe(apiAddr, mux); err != nil && !errors.Is(err, http.ErrServerClosed) {
		sugarzero.Errorf(ctx, "http server stopped: %v", err)
	}
}

func logLevelHandler(ctx context.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reqCtx := sugarzero.WithFields(ctx,
			"remote_addr", r.RemoteAddr,
			"method", r.Method,
		)

		switch r.Method {
		case http.MethodGet:
			writeJSON(w, http.StatusOK, newLevelResponse(reqCtx))
		case http.MethodPost:
			var payload levelChangeRequest
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body"})
				return
			}

			desired := strings.ToLower(strings.TrimSpace(payload.Level))
			if desired == "" {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "level is required"})
				return
			}

			if _, ok := allowedLookup[desired]; !ok {
				writeJSON(w, http.StatusBadRequest, errorResponse{Error: "level must be one of " + strings.Join(allowedLevels, ", ")})
				return
			}

			if err := sugarzero.SetLogLevel(reqCtx, desired); err != nil {
				writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
				return
			}

			sugarzero.Infof(reqCtx, "log level updated to %s", desired)
			writeJSON(w, http.StatusOK, newLevelResponse(reqCtx))
		default:
			w.Header().Set("Allow", "GET, POST")
			writeJSON(w, http.StatusMethodNotAllowed, errorResponse{Error: "method not allowed"})
		}
	}
}

func emitBackgroundLogs(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		sugarzero.Debug(ctx, "background heartbeat running")
		sugarzero.Info(ctx, "service healthy, adjust log level via POST /log-level")
		sugarzero.Warn(ctx, "set level back to info to reduce verbose output when done")
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func newLevelResponse(ctx context.Context) levelResponse {
	return levelResponse{
		CurrentLevel: sugarzero.GetLogLevel(ctx),
		DefaultLevel: defaultLevel,
		Allowed:      allowedLevels,
	}
}
