import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { listLocalExecutors, localExecutorStatus } from '../api/localExecutors.js';

export function createLocalExecutorRefreshGate(load) {
  let pending = null;
  return {
    refresh() {
      if (!pending) pending = Promise.resolve(load()).finally(() => { pending = null; });
      return pending;
    }
  };
}

export function useLocalExecutorStatus({ enabled = true, intervalMs = 30000 } = {}) {
  const [executors, setExecutors] = useState([]);
  const [loading, setLoading] = useState(Boolean(enabled));
  const [error, setError] = useState(null);
  const gateRef = useRef(null);

  if (!gateRef.current) {
    gateRef.current = createLocalExecutorRefreshGate(async () => {
      setLoading(true);
      try {
        const next = await listLocalExecutors();
        setExecutors(next);
        setError(null);
        return next;
      } catch (nextError) {
        setError(nextError);
        throw nextError;
      } finally {
        setLoading(false);
      }
    });
  }

  const refresh = useCallback(() => gateRef.current.refresh(), []);

  useEffect(() => {
    if (!enabled) return undefined;
    let timer = null;
    const start = () => {
      void refresh().catch(() => {});
      timer = window.setInterval(() => { void refresh().catch(() => {}); }, intervalMs);
    };
    const onVisibilityChange = () => {
      if (document.hidden) {
        if (timer) window.clearInterval(timer);
        timer = null;
      } else if (!timer) {
        start();
      }
    };
    if (!document.hidden) start();
    document.addEventListener('visibilitychange', onVisibilityChange);
    return () => {
      if (timer) window.clearInterval(timer);
      document.removeEventListener('visibilitychange', onVisibilityChange);
    };
  }, [enabled, intervalMs, refresh]);

  return { executors, status: useMemo(() => localExecutorStatus(executors), [executors]), loading, error, refresh };
}
