package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"github.com/akz/nexus-dashboard/backend/internal/config"
	"github.com/akz/nexus-dashboard/backend/internal/database"
	"github.com/akz/nexus-dashboard/backend/internal/models"
	"github.com/akz/nexus-dashboard/backend/internal/repository"
	"github.com/akz/nexus-dashboard/backend/pkg/logger"
)

// Design Rule #3: Decouple Reads from Writes via Contexts.
// The API server runs independently from the ingestor process.
// All queries use context.WithTimeout with a strict 2-second cutoff.

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info("starting nexus api server", "port", cfg.APIPort, "query_timeout", cfg.APIQueryTimeout)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---------- Database ----------
	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxOpen, cfg.DBMaxIdle, cfg.DBMaxLifetime)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	metricRepo := repository.NewMetricRepository(pool, log)

	// ---------- WebSocket upgrader ----------
	wsUpgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // allow all origins in dev
		},
	}

	// ---------- HTTP router ----------
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(cfg.APIQueryTimeout))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := metricRepo.Ping(ctx); err != nil {
			http.Error(w, `{"status":"down"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	// REST API: query historical metric summaries
	r.Get("/api/metrics/summary", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		since, err := time.Parse(time.RFC3339, q.Get("since"))
		if err != nil {
			since = time.Now().Add(-1 * time.Hour)
		}
		until, err := time.Parse(time.RFC3339, q.Get("until"))
		if err != nil {
			until = time.Now()
		}

		filter := models.MetricFilter{
			Since:    since,
			Until:    until,
			DeviceID: q.Get("device_id"),
			Limit:    1000,
			Offset:   0,
		}

		// 2-second query timeout per Design Rule #3
		queryCtx, cancel := context.WithTimeout(r.Context(), cfg.APIQueryTimeout)
		defer cancel()

		results, err := metricRepo.QuerySummary(queryCtx, filter)
		if err != nil {
			log.Error("summary query failed", "error", err)
			http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":  results,
			"count": len(results),
		})
	})

	// REST API: raw data points
	r.Get("/api/metrics/raw", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		since, err := time.Parse(time.RFC3339, q.Get("since"))
		if err != nil {
			since = time.Now().Add(-1 * time.Hour)
		}
		until, err := time.Parse(time.RFC3339, q.Get("until"))
		if err != nil {
			until = time.Now()
		}

		filter := models.MetricFilter{
			Since:    since,
			Until:    until,
			DeviceID: q.Get("device_id"),
			Limit:    500,
		}

		queryCtx, cancel := context.WithTimeout(r.Context(), cfg.APIQueryTimeout)
		defer cancel()

		results, err := metricRepo.QueryRaw(queryCtx, filter)
		if err != nil {
			log.Error("raw query failed", "error", err)
			http.Error(w, `{"error":"query failed"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":  results,
			"count": len(results),
		})
	})

	// WebSocket endpoint for real-time streaming
	r.Get("/ws/metrics", func(w http.ResponseWriter, r *http.Request) {
		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("websocket upgrade failed", "error", err)
			return
		}
		defer conn.Close()

		log.Info("websocket client connected", "remote", r.RemoteAddr)

		// Keep connection alive — ingestor pushes data through DB,
		// clients poll or wait. For simplicity, send a heartbeat.
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		done := make(chan struct{})
		go func() {
			defer close(done)
			for {
				_, _, err := conn.ReadMessage()
				if err != nil {
					return
				}
			}
		}()

		for {
			select {
			case <-done:
				log.Info("websocket client disconnected", "remote", r.RemoteAddr)
				return
			case <-ticker.C:
				if err := conn.WriteJSON(map[string]string{"type": "heartbeat"}); err != nil {
					return
				}
			}
		}
	})

	// ---------- Server ----------
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.APIPort),
		Handler:      r,
		ReadTimeout:  cfg.APIReadTimeout,
		WriteTimeout: cfg.APIWriteTimeout,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info("api server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", "error", err)
		}
	}()

	<-sigCh
	log.Info("shutting down api server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
	wg.Wait()
	log.Info("api server shutdown complete")
}
