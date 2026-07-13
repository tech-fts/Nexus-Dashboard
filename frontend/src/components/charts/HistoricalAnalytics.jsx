import React, { useMemo } from 'react';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { formatDate } from '../../utils/formatters';

/**
 * HistoricalAnalytics — bar chart of pre-aggregated metric summaries.
 *
 * Queries the continuous aggregate views in TimescaleDB (metric_summary_hourly)
 * via the Go API. Since data is pre-aggregated, both query (<50ms) and rendering
 * are fast regardless of how many raw data points exist.
 *
 * Design Rule #2 (TimescaleDB): Queries continuous aggregates, not raw tables.
 */
export default function HistoricalAnalytics({ data, isLoading }) {
  const chartData = useMemo(() => {
    if (!data || data.length === 0) return [];

    // Group by bucket and average
    const buckets = {};
    for (const d of data) {
      const key = formatDate(d.bucket);
      if (!buckets[key]) {
        buckets[key] = { bucket: key, total: 0, count: 0 };
      }
      buckets[key].total += d.avg_value || 0;
      buckets[key].count += 1;
    }

    return Object.values(buckets).map((b) => ({
      bucket: b.bucket,
      avg_value: b.count > 0 ? b.total / b.count : 0,
    }));
  }, [data]);

  if (isLoading) {
    return (
      <div className="bg-gray-900 border border-gray-800 rounded-lg p-8">
        <h2 className="text-lg font-medium text-gray-300 mb-2">Historical Analytics</h2>
        <p className="text-gray-500 text-sm">Loading aggregate data...</p>
      </div>
    );
  }

  if (!data || data.length === 0) {
    return (
      <div className="bg-gray-900 border border-gray-800 rounded-lg p-8">
        <h2 className="text-lg font-medium text-gray-300 mb-2">Historical Analytics</h2>
        <p className="text-gray-500 text-sm">No historical data available</p>
      </div>
    );
  }

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-lg p-6">
      <h2 className="text-lg font-medium text-gray-300 mb-4">
        Historical Analytics
      </h2>
      <p className="text-xs text-gray-500 mb-4">
        Pre-aggregated hourly averages (TimescaleDB continuous aggregate)
      </p>
      <div className="h-[400px]">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
            <XAxis
              dataKey="bucket"
              stroke="#64748b"
              tick={{ fill: '#64748b', fontSize: 11 }}
              tickLine={false}
              interval="preserveStartEnd"
            />
            <YAxis
              stroke="#64748b"
              tick={{ fill: '#64748b', fontSize: 11 }}
              tickLine={false}
              axisLine={false}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: '#1e293b',
                border: '1px solid #334155',
                borderRadius: '6px',
                color: '#e2e8f0',
                fontSize: '12px',
              }}
            />
            <Legend
              wrapperStyle={{ fontSize: '12px', color: '#94a3b8' }}
            />
            <Bar
              dataKey="avg_value"
              fill="#5c7cfa"
              radius={[4, 4, 0, 0]}
              name="Avg Value"
            />
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
