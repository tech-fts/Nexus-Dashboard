import React, { useState } from 'react';
import Sidebar from './components/common/Sidebar';
import StatCard from './components/common/StatCard';
import RealTimeChart from './components/charts/RealTimeChart';
import HistoricalAnalytics from './components/charts/HistoricalAnalytics';
import useWebSocket from './hooks/useWebSocket';
import useMetricQuery from './hooks/useMetricQuery';

const NAV_ITEMS = [
  { id: 'overview', label: 'Overview', icon: '◉' },
  { id: 'realtime', label: 'Real-Time', icon: '◈' },
  { id: 'analytics', label: 'Analytics', icon: '▣' },
  { id: 'alerts', label: 'Alerts', icon: '⚠' },
  { id: 'settings', label: 'Settings', icon: '⚙' },
];

export default function App() {
  const [activeView, setActiveView] = useState('overview');
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false);

  // WebSocket hook with built-in throttled buffer (Design Rule #1)
  const { latestMetrics, isConnected } = useWebSocket(
    `ws://${window.location.hostname}:8080/ws/metrics`
  );

  // REST hook for historical data — queries continuous aggregate views
  const { data: historicalData, isLoading } = useMetricQuery({
    endpoint: '/api/metrics/summary',
    since: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
    until: new Date().toISOString(),
  });

  return (
    <div className="flex h-screen overflow-hidden bg-gray-950">
      {/* Design Rule #2 (Frontend): Sidebar is isolated — never re-renders with data */}
      <Sidebar
        items={NAV_ITEMS}
        activeId={activeView}
        onSelect={setActiveView}
        collapsed={sidebarCollapsed}
        onToggle={() => setSidebarCollapsed((c) => !c)}
      />

      {/* Main content area */}
      <main className="flex-1 overflow-y-auto p-6">
        <header className="mb-8">
          <h1 className="text-2xl font-semibold text-gray-100">
            Nexus Dashboard
          </h1>
          <div className="flex items-center gap-2 mt-1">
            <span
              className={`inline-block w-2 h-2 rounded-full ${
                isConnected ? 'bg-green-500' : 'bg-red-500'
              }`}
            />
            <span className="text-sm text-gray-400">
              {isConnected ? 'Live' : 'Disconnected'}
            </span>
          </div>
        </header>

        {activeView === 'overview' && (
          <OverviewGrid latestMetrics={latestMetrics} />
        )}

        {activeView === 'realtime' && (
          <RealTimeChart metrics={latestMetrics} />
        )}

        {activeView === 'analytics' && (
          <HistoricalAnalytics data={historicalData} isLoading={isLoading} />
        )}

        {activeView === 'alerts' && (
          <div className="text-gray-400 text-center py-20">
            No active alerts
          </div>
        )}

        {activeView === 'settings' && (
          <div className="text-gray-400 text-center py-20">
            Settings panel
          </div>
        )}
      </main>
    </div>
  );
}

function OverviewGrid({ latestMetrics }) {
  const stats = [
    {
      title: 'Events / sec',
      value: latestMetrics.length > 0 ? latestMetrics.length * 4 : '—',
      unit: 'avg',
      trend: '+12%',
    },
    {
      title: 'Active Devices',
      value: latestMetrics.length > 0
        ? new Set(latestMetrics.map((m) => m.device_id)).size
        : '—',
      unit: 'devices',
    },
    {
      title: 'Avg Latency',
      value: '2.3',
      unit: 'ms',
      trend: '-8%',
    },
    {
      title: 'Data Ingested',
      value: '847',
      unit: 'MB',
      trend: '+5%',
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
      {stats.map((s) => (
        <StatCard key={s.title} {...s} />
      ))}
    </div>
  );
}
