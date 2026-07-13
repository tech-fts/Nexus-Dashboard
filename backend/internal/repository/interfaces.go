package repository

import (
	"context"
	"time"

	"github.com/akz/nexus-dashboard/backend/internal/models"
)

// MetricRepository defines the data access contract for metrics.
// This interface decouples DB implementation for testing (SOLID: Dependency Inversion).
type MetricRepository interface {
	// BatchInsert flushes a batch of metrics using a single multi-row insert.
	// Design Rule #1 (Ingestor): Never insert one row at a time.
	BatchInsert(ctx context.Context, metrics []*models.Metric) error

	// QuerySummary retrieves pre-aggregated data from continuous aggregate views.
	// Design Rule #2 (TimescaleDB): Query materialized views, not raw hypertables.
	QuerySummary(ctx context.Context, filter models.MetricFilter) ([]*models.MetricSummary, error)

	// QueryRaw retrieves raw metric data points within a time range.
	QueryRaw(ctx context.Context, filter models.MetricFilter) ([]*models.Metric, error)

	// Ping checks database connectivity.
	Ping(ctx context.Context) error

	// Close releases any held resources.
	Close()
}
