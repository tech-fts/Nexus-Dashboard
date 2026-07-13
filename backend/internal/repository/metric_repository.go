package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/akz/nexus-dashboard/backend/internal/models"
	"github.com/akz/nexus-dashboard/backend/pkg/logger"
)

type metricRepository struct {
	pool *pgxpool.Pool
	log  *logger.Logger
}

// NewMetricRepository creates a concrete MetricRepository backed by pgxpool.
func NewMetricRepository(pool *pgxpool.Pool, log *logger.Logger) MetricRepository {
	return &metricRepository{pool: pool, log: log}
}

// BatchInsert performs a single multi-row INSERT for a batch of metrics.
// Design Rule #1 (Ingestor): Batch inserts, never one-by-one.
func (r *metricRepository) BatchInsert(ctx context.Context, metrics []*models.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// Use pgx's CopyFrom for optimal batch performance
	rows := make([][]interface{}, len(metrics))
	for i, m := range metrics {
		tagsJSON := "{}"
		if m.Tags != nil {
			tagsJSON = mapToJSON(m.Tags)
		}
		rows[i] = []interface{}{
			m.ID,
			m.DeviceID,
			m.Name,
			m.Value,
			m.Unit,
			tagsJSON,
			m.Timestamp,
		}
	}

	_, err := r.pool.CopyFrom(
		ctx,
		[]string{"metrics"},
		[]string{"id", "device_id", "name", "value", "unit", "tags", "time"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return fmt.Errorf("batch insert %d metrics: %w", len(metrics), err)
	}

	r.log.Debug("batch inserted metrics", "count", len(metrics))
	return nil
}

// QuerySummary queries the continuous aggregate materialized view.
// Design Rule #2 (TimescaleDB): React queries pre-computed views, not raw data.
func (r *metricRepository) QuerySummary(ctx context.Context, filter models.MetricFilter) ([]*models.MetricSummary, error) {
	// Use the hourly continuous aggregate view
	query := `SELECT
		bucket,
		name,
		avg_value,
		max_value,
		min_value,
		count
	FROM metric_summary_hourly
	WHERE bucket >= $1 AND bucket < $2`

	args := []interface{}{filter.Since, filter.Until}
	argIdx := 3

	if filter.DeviceID != "" {
		query += fmt.Sprintf(" AND device_id = $%d", argIdx)
		args = append(args, filter.DeviceID)
		argIdx++
	}

	if len(filter.Names) > 0 {
		placeholders := make([]string, len(filter.Names))
		for i, name := range filter.Names {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, name)
			argIdx++
		}
		query += " AND name IN (" + strings.Join(placeholders, ",") + ")"
	}

	query += " ORDER BY bucket ASC, name ASC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argIdx)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query summary: %w", err)
	}
	defer rows.Close()

	var results []*models.MetricSummary
	for rows.Next() {
		s := &models.MetricSummary{}
		if err := rows.Scan(&s.Bucket, &s.Name, &s.AvgValue, &s.MaxValue, &s.MinValue, &s.Count); err != nil {
			return nil, fmt.Errorf("scan summary row: %w", err)
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// QueryRaw retrieves raw metric rows within a time range.
func (r *metricRepository) QueryRaw(ctx context.Context, filter models.MetricFilter) ([]*models.Metric, error) {
	query := `SELECT id, device_id, name, value, unit, tags, time
	FROM metrics
	WHERE time >= $1 AND time < $2`

	args := []interface{}{filter.Since, filter.Until}
	argIdx := 3

	if filter.DeviceID != "" {
		query += fmt.Sprintf(" AND device_id = $%d", argIdx)
		args = append(args, filter.DeviceID)
		argIdx++
	}

	if len(filter.Names) > 0 {
		placeholders := make([]string, len(filter.Names))
		for i, name := range filter.Names {
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, name)
			argIdx++
		}
		query += " AND name IN (" + strings.Join(placeholders, ",") + ")"
	}

	query += " ORDER BY time DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argIdx)
		args = append(args, filter.Limit)
		argIdx++
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query raw: %w", err)
	}
	defer rows.Close()

	var results []*models.Metric
	for rows.Next() {
		m := &models.Metric{}
		var tagsJSON *string
		if err := rows.Scan(&m.ID, &m.DeviceID, &m.Name, &m.Value, &m.Unit, &tagsJSON, &m.Timestamp); err != nil {
			return nil, fmt.Errorf("scan metric row: %w", err)
		}
		if tagsJSON != nil {
			m.Tags = parseJSONMap(*tagsJSON)
		}
		results = append(results, m)
	}
	return results, rows.Err()
}

func (r *metricRepository) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

func (r *metricRepository) Close() {
	r.pool.Close()
}

// --- helpers ---

// mapToJSON converts a map to a JSON string for storage in JSONB column.
func mapToJSON(m map[string]string) string {
	if len(m) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteByte('{')
	first := true
	for k, v := range m {
		if !first {
			b.WriteByte(',')
		}
		b.WriteString(`"`)
		b.WriteString(k)
		b.WriteString(`":"`)
		b.WriteString(v)
		b.WriteString(`"`)
		first = false
	}
	b.WriteByte('}')
	return b.String()
}

// parseJSONMap is a placeholder — in production use json.Unmarshal.
func parseJSONMap(s string) map[string]string {
	m := make(map[string]string)
	// simplified: real impl should use encoding/json
	_ = s
	return m
}
