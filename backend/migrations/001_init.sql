-- ============================================================
-- Nexus Dashboard — TimescaleDB Schema Migration
-- ============================================================
-- Design rules enforced:
--   1. Hypertable with time as primary dimension (no SERIAL PK)
--   2. Continuous aggregate materialized views for fast queries
--   3. JSONB tags for flexible metric metadata
-- ============================================================

-- Enable the TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb WITH SCHEMA public;

-- ============================================================
-- Raw metrics hypertable
-- ============================================================
-- Design Rule #1: No auto-incrementing SERIAL PK for time-series.
-- Use UUID + timestamp as the natural key.
CREATE TABLE IF NOT EXISTS metrics (
    id          TEXT        NOT NULL,
    device_id   TEXT        NOT NULL,
    name        TEXT        NOT NULL,
    value       DOUBLE PRECISION NOT NULL,
    unit        TEXT        DEFAULT '',
    tags        JSONB       DEFAULT '{}',
    time        TIMESTAMPTZ NOT NULL,

    -- Composite key: device + name + time ensures uniqueness
    PRIMARY KEY (device_id, name, time)
);

-- Convert to hypertable partitioned on 'time'
SELECT create_hypertable('metrics', 'time', if_not_exists => TRUE);

-- Index for efficient device-scoped queries
CREATE INDEX IF NOT EXISTS idx_metrics_device_time
    ON metrics (device_id, time DESC);

-- Index for name-based queries (used by continuous aggregate)
CREATE INDEX IF NOT EXISTS idx_metrics_name_time
    ON metrics (name, time DESC);

-- ============================================================
-- Continuous Aggregate: Hourly metric summaries
-- ============================================================
-- Design Rule #2: Pre-calculate trends in the background.
-- React queries this view, never the raw metrics table.
-- This keeps UI queries under 50ms regardless of data volume.
CREATE MATERIALIZED VIEW IF NOT EXISTS metric_summary_hourly
WITH (timescaledb.continuous) AS
SELECT
    time_bucket('1 hour', time) AS bucket,
    device_id,
    name,
    AVG(value) AS avg_value,
    MAX(value) AS max_value,
    MIN(value) AS min_value,
    COUNT(*)   AS count
FROM metrics
GROUP BY bucket, device_id, name
WITH NO DATA;

-- Add an index on the continuous aggregate for fast lookups
CREATE INDEX IF NOT EXISTS idx_summary_hourly_bucket
    ON metric_summary_hourly (bucket DESC, device_id, name);

-- Refresh policy: materialize new data every hour
SELECT add_continuous_aggregate_policy('metric_summary_hourly',
    start_offset    => INTERVAL '2 days',
    end_offset      => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour',
    if_not_exists   => TRUE
);

-- ============================================================
-- Retention policy: drop raw data older than 30 days
-- ============================================================
SELECT add_retention_policy('metrics', INTERVAL '30 days', if_not_exists => TRUE);
