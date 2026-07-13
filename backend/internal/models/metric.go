package models

import (
	"time"
)

// Metric represents a single telemetry data point ingested from Kafka.
// Fields use JSON tags for API serialization and DB tags for SQL mapping.
type Metric struct {
	ID        string    `json:"id" db:"id"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	Name      string    `json:"name" db:"name"`
	Value     float64   `json:"value" db:"value"`
	Unit      string    `json:"unit,omitempty" db:"unit"`
	Tags      map[string]string `json:"tags,omitempty" db:"tags"`
	Timestamp time.Time `json:"timestamp" db:"time"`
}

// MetricSummary is the pre-aggregated row returned from continuous aggregates.
// Design Rule #2 (TimescaleDB): React queries these views, not raw tables.
type MetricSummary struct {
	Bucket    time.Time `json:"bucket"`
	Name      string    `json:"name"`
	AvgValue  float64   `json:"avg_value"`
	MaxValue  float64   `json:"max_value"`
	MinValue  float64   `json:"min_value"`
	Count     int64     `json:"count"`
	DeviceID  string    `json:"device_id,omitempty"`
}

// MetricFilter carries query parameters for the REST API.
type MetricFilter struct {
	Names    []string  `json:"names,omitempty"`
	DeviceID string    `json:"device_id,omitempty"`
	Since    time.Time `json:"since"`
	Until    time.Time `json:"until"`
	Bucket   string    `json:"bucket"` // "1 hour", "1 day" etc
	Limit    int       `json:"limit,omitempty"`
	Offset   int       `json:"offset,omitempty"`
}
