import { useState, useEffect, useCallback } from 'react';

/**
 * useMetricQuery — REST data fetching with loading/cancellation.
 *
 * Queries the Go API's continuous aggregate endpoints.
 * Design Rule #2 (TimescaleDB): React queries pre-computed aggregate views,
 * not raw hypertables, keeping response times under 50ms.
 */
export default function useMetricQuery({ endpoint, since, until, deviceId, names }) {
  const [data, setData] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);

  const fetchData = useCallback(async (signal) => {
    setIsLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      if (since) params.set('since', since);
      if (until) params.set('until', until);
      if (deviceId) params.set('device_id', deviceId);
      if (names?.length) params.set('names', names.join(','));

      const url = `${endpoint}?${params.toString()}`;
      const res = await fetch(url, { signal });

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}: ${res.statusText}`);
      }

      const json = await res.json();
      setData(json.data || []);
    } catch (err) {
      if (err.name === 'AbortError') return;
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  }, [endpoint, since, until, deviceId, names]);

  useEffect(() => {
    const controller = new AbortController();
    fetchData(controller.signal);
    return () => controller.abort();
  }, [fetchData]);

  return { data, isLoading, error, refetch: () => fetchData(new AbortController().signal) };
}
