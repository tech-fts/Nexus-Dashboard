import React, { useMemo } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';
import { formatTimestamp } from '../../utils/formatters';

/**
 * RealTimeChart — live streaming line chart.
 * Design Rule #2 (Frontend): Data flows directly to this leaf component.
 * Parent wrappers (header, sidebar) are isolated from re-renders.
 *
 * Design Rule #1 (Frontend): The data is already throttled by useWebSocket's
 * 250ms buffer — this chart only re-renders at the flush rate.
 */
export default function RealTimeChart({ metrics }) {
  const chartData = useMemo(() => {
    if (!metrics || metrics.length === 0) return [];

    // Group by name for multiple series
    const grouped = {};
    for (const m of metrics) {
      if (!grouped[m.name]) grouped[m.name] = [];
      grouped[m.name].push({
        time: formatTimestamp(m.timestamp),
        value: m.value,
      });
    }

    // Take the first metric name for single-series display
    const names = Object.keys(grouped);
    return names.length > 0
      ? grouped[names[0]].slice(-50)
      : [];
  }, [metrics]);

  if (!metrics || metrics.length === 0) {
    return (
      <div className="bg-gray-900 border border-gray-800 rounded-lg p-8">
        <h2 className="text-lg font-medium text-gray-300 mb-2">Real-Time Metrics</h2>
        <p className="text-gray-500 text-sm">Waiting for data stream...</p>
      </div>
    );
  }

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-lg p-6">
      <h2 className="text-lg font-medium text-gray-300 mb-4">Real-Time Metrics</h2>
      <div className="h-[400px]">
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#1e293b" />
            <XAxis
              dataKey="time"
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
            <Line
              type="monotone"
              dataKey="value"
              stroke="#5c7cfa"
              strokeWidth={2}
              dot={false}
              activeDot={{ r: 4, fill: '#5c7cfa' }}
              isAnimationActive={false}
            />
          </LineChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
