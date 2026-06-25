// Command api is the OfficeHours.ai HTTP backend.
//
// This is the compiling skeleton: it serves /api/health, wires an (otherwise
// empty) router under /api, and starts a job-worker goroutine stub that other
// agents will flesh out. Routes from BUILD_SPEC §4 are registered by the
// handlers package as it is filled in.
package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"officehours/internal/config"
	"officehours/internal/db"
	"officehours/internal/handlers"
	"officehours/internal/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Printf("api: config load failed (continuing with env-only): %v", err)
		cfg = &config.Config{Port: getenv("PORT", "8080")}
	}

	// Connect to Postgres. Non-fatal at boot so the container can come up and
	// report health while the db warms; handlers can re-check as needed.
	database, err := db.ConnectFromEnv(ctx)
	if err != nil {
		log.Printf("api: database connect failed (will serve health only): %v", err)
	} else {
		defer database.Close()
		log.Printf("api: connected to database")
		// Apply the frozen schema on boot (idempotent: skips if already present).
		if err := database.Migrate(ctx); err != nil {
			log.Printf("api: migration failed: %v", err)
		}
	}

	mux := http.NewServeMux()

	// API router. Feature agents register their routes here.
	api := http.NewServeMux()
	api.HandleFunc("GET /health", healthHandler(database))
	registerRoutes(api, cfg, database)

	mux.Handle("/api/", http.StripPrefix("/api", api))

	// Start the background job worker (stub).
	go runJobWorker(ctx, cfg, database)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("api: listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("api: server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Printf("api: shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

// registerRoutes wires the handlers package onto the /api mux.
func registerRoutes(api *http.ServeMux, cfg *config.Config, database *db.DB) {
	handlers.Register(api, cfg, database)
}

// healthHandler reports liveness and db reachability.
func healthHandler(database *db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "ok"
		dbOK := false
		if database != nil && database.Pool != nil {
			pingCtx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
			defer cancel()
			if err := database.Pool.Ping(pingCtx); err == nil {
				dbOK = true
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status": status,
			"db":     dbOK,
		})
	}
}

// runJobWorker polls agent_jobs and execs claude per job (see internal/worker).
func runJobWorker(ctx context.Context, cfg *config.Config, database *db.DB) {
	w := &worker.Worker{
		Cfg:      cfg,
		DB:       database,
		Workers:  workerCount(),
		OhctlDir: getenv("OHCTL_DIR", "/usr/local/bin"),
	}
	w.Run(ctx)
}

// workerCount reads JOB_WORKERS or defaults to 2.
func workerCount() int {
	if v := os.Getenv("JOB_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return n
		}
	}
	return 2
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
