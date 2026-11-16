import { useEffect, useRef } from 'react';

// The render-time epoch closes the gap before effect cleanup on a fast switch.
export default function useWorkspaceScope(key) {
  const ref = useRef();
  if (!ref.current || ref.current.key !== key) {
    ref.current?.controller.abort();
    ref.current = { key, controller: new AbortController() };
  }
  const entry = ref.current;
  useEffect(() => () => entry.controller.abort(), [entry]);
  return {
    signal: entry.controller.signal,
    isCurrent: () => ref.current === entry && !entry.controller.signal.aborted,
    async call(promise, apply) {
      try { const result = await promise; if (ref.current === entry && !entry.controller.signal.aborted) { apply?.(result); return result; } }
      catch (error) { if (ref.current === entry && !entry.controller.signal.aborted) throw error; }
    }
  };
}
