import React from 'react';

/**
 * Sidebar — navigation shell.
 * Design Rule #2 (Frontend): Isolated from data streams.
 * Never re-renders when metrics arrive because it receives no changing props.
 */
export default function Sidebar({ items, activeId, onSelect, collapsed, onToggle }) {
  return (
    <aside
      className={`flex flex-col bg-gray-900 border-r border-gray-800 transition-all duration-200 ${
        collapsed ? 'w-16' : 'w-56'
      }`}
    >
      {/* Logo / Toggle */}
      <div className="flex items-center justify-between h-14 px-4 border-b border-gray-800">
        {!collapsed && (
          <span className="text-lg font-semibold text-nexus-400 tracking-wide">
            NEXUS
          </span>
        )}
        <button
          onClick={onToggle}
          className="text-gray-400 hover:text-gray-200 transition-colors p-1"
          aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
        >
          {collapsed ? '→' : '←'}
        </button>
      </div>

      {/* Navigation items */}
      <nav className="flex-1 py-4">
        {items.map((item) => (
          <button
            key={item.id}
            onClick={() => onSelect(item.id)}
            className={`flex items-center gap-3 w-full px-4 py-2.5 text-sm transition-colors ${
              activeId === item.id
                ? 'text-nexus-400 bg-nexus-900/30 border-r-2 border-nexus-500'
                : 'text-gray-400 hover:text-gray-200 hover:bg-gray-800/50'
            }`}
            title={collapsed ? item.label : undefined}
          >
            <span className="text-lg w-5 text-center flex-shrink-0">{item.icon}</span>
            {!collapsed && <span className="truncate">{item.label}</span>}
          </button>
        ))}
      </nav>

      {/* Connection status footer */}
      {!collapsed && (
        <div className="px-4 py-3 border-t border-gray-800 text-xs text-gray-500">
          Nexus v1.0
        </div>
      )}
    </aside>
  );
}
