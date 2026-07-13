import { useEffect, useRef, useState, useCallback } from 'react';

/**
 * useWebSocket — Real-time event stream with a throttled rendering buffer.
 *
 * Design Rule #1 (Frontend): Throttle State Updates.
 * Incoming WebSocket events accumulate in a mutable useRef array and flush
 * into visible useState at a controlled rate (default 250ms), preventing
 * layout thrashing from render-on-every-event.
 */
export default function useWebSocket(url, throttleMs = 250) {
  const [latestMetrics, setLatestMetrics] = useState([]);
  const [isConnected, setIsConnected] = useState(false);
  const bufferRef = useRef([]);
  const wsRef = useRef(null);
  const flushIdRef = useRef(null);
  const reconnectTimeoutRef = useRef(null);

  const flush = useCallback(() => {
    if (bufferRef.current.length === 0) return;

    // Swap the buffer atomically — minimal GC pressure
    const batch = bufferRef.current;
    bufferRef.current = [];

    setLatestMetrics((prev) => {
      // Keep at most 200 data points in visible state
      const merged = [...prev, ...batch];
      return merged.length > 200 ? merged.slice(-200) : merged;
    });
  }, []);

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return;

    const ws = new WebSocket(url);
    wsRef.current = ws;

    ws.onopen = () => {
      setIsConnected(true);
      // Start throttled flush cycle
      flushIdRef.current = setInterval(flush, throttleMs);
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        // Skip heartbeats
        if (data.type === 'heartbeat') return;

        // Push into mutable buffer — no re-render
        bufferRef.current.push(data);
      } catch {
        // ignore malformed frames
      }
    };

    ws.onclose = () => {
      setIsConnected(false);
      clearInterval(flushIdRef.current);
      wsRef.current = null;

      // Auto-reconnect after 3 seconds
      reconnectTimeoutRef.current = setTimeout(connect, 3000);
    };

    ws.onerror = () => {
      ws.close();
    };
  }, [url, throttleMs, flush]);

  useEffect(() => {
    connect();
    return () => {
      clearInterval(flushIdRef.current);
      clearTimeout(reconnectTimeoutRef.current);
      wsRef.current?.close();
    };
  }, [connect]);

  return { latestMetrics, isConnected };
}
