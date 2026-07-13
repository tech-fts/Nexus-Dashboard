package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/akz/nexus-dashboard/backend/internal/config"
	"github.com/akz/nexus-dashboard/backend/internal/database"
	"github.com/akz/nexus-dashboard/backend/internal/models"
	"github.com/akz/nexus-dashboard/backend/internal/queue"
	"github.com/akz/nexus-dashboard/backend/internal/repository"
	"github.com/akz/nexus-dashboard/backend/pkg/logger"
)

// Design Rule #1 (Ingestor): The Batching Buffer Pattern.
// Kafka events are queued into a memory channel and flushed to TimescaleDB
// via batch COPY every BatchSize events or BatchFlushInterval, whichever comes first.

func main() {
	// Load .env if present
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		panic("config: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info("starting nexus ingestor", "batch_size", cfg.BatchSize, "flush_interval", cfg.BatchFlushInterval)

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// ---------- Database connection ----------
	pool, err := database.Connect(ctx, cfg.DatabaseURL, cfg.DBMaxOpen, cfg.DBMaxIdle, cfg.DBMaxLifetime)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()
	log.Info("connected to timescaledb")

	// ---------- Repository ----------
	metricRepo := repository.NewMetricRepository(pool, log)

	// ---------- Kafka consumer ----------
	consumer := queue.NewConsumer(cfg.KafkaBrokers, cfg.KafkaTopic, cfg.KafkaGroupID, log)
	defer consumer.Close()

	// ---------- Batching buffer ----------
	// Metrics are collected in a buffered channel and flushed in batches.
	batchCh := make(chan *metricBatch, 1)
	flushDone := make(chan struct{})

	// Worker goroutine: flushes batches to TimescaleDB
	go flushWorker(ctx, metricRepo, batchCh, flushDone, log)

	// Read loop: collect events into batches and send to flush worker
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		defer close(batchCh)

		batch := &metricBatch{metrics: make([]*models.Metric, 0, cfg.BatchSize)}
		flushTicker := time.NewTicker(cfg.BatchFlushInterval)
		defer flushTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				// Flush remaining on shutdown
				if len(batch.metrics) > 0 {
					batchCh <- batch
				}
				return

			case <-flushTicker.C:
				// Time-based flush: 500ms threshold
				if len(batch.metrics) > 0 {
					batchCh <- batch
					batch = &metricBatch{metrics: make([]*models.Metric, 0, cfg.BatchSize)}
				}

			default:
				msg, err := consumer.ReadMessage(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return // shutting down
					}
					log.Error("kafka read error", "error", err)
					continue
				}

				metric, err := queue.DecodeMetric(msg)
				if err != nil {
					log.Warn("failed to decode metric", "error", err, "offset", msg.Offset)
					continue
				}

				batch.metrics = append(batch.metrics, metric)

				// Size-based flush: 1000 events threshold
				if len(batch.metrics) >= cfg.BatchSize {
					batchCh <- batch
					batch = &metricBatch{metrics: make([]*models.Metric, 0, cfg.BatchSize)}
					flushTicker.Reset(cfg.BatchFlushInterval)
				}
			}
		}
	}()

	// Wait for shutdown signal
	<-sigCh
	log.Info("shutdown signal received, draining...")
	cancel()    // stop read loop
	wg.Wait()   // wait for read loop to finish
	<-flushDone // wait for last flush to complete
	log.Info("ingestor shutdown complete")
}

// --- internal types ---

type metricBatch struct {
	metrics []*models.Metric
}

func flushWorker(
	ctx context.Context,
	repo repository.MetricRepository,
	batchCh <-chan *metricBatch,
	done chan<- struct{},
	log *logger.Logger,
) {
	defer close(done)

	for batch := range batchCh {
		if len(batch.metrics) == 0 {
			continue
		}

		insertCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		err := repo.BatchInsert(insertCtx, batch.metrics)
		cancel()

		if err != nil {
			log.Error("batch insert failed", "error", err, "count", len(batch.metrics))
			continue
		}

		log.Debug("batch flushed", "count", len(batch.metrics))
	}
}
