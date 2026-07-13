import React from 'react';

/**
 * StatCard — single KPI display card.
 * Receives simple scalar props so it's a pure leaf component.
 * Design Rule #2 (Frontend): Data streams to leaf nodes — parents don't re-render.
 */
export default function StatCard({ title, value, unit, trend }) {
  const trendUp = trend?.startsWith('+');

  return (
    <div className="bg-gray-900 border border-gray-800 rounded-lg p-4 hover:border-gray-700 transition-colors">
      <div className="text-xs font-medium text-gray-500 uppercase tracking-wider mb-1">
        {title}
      </div>
      <div className="flex items-baseline gap-1.5">
        <span className="text-2xl font-semibold text-gray-100">{value}</span>
        {unit && <span className="text-xs text-gray-400">{unit}</span>}
      </div>
      {trend && (
        <div
          className={`mt-1 text-xs font-medium ${
            trendUp ? 'text-green-400' : 'text-red-400'
          }`}
        >
          {trend} vs last hour
        </div>
      )}
    </div>
  );
}
