export function createDirectorSaveQueue(write, { onStatus = () => {} } = {}) {
  let pending = null;
  let inFlight = null;
  let status = 'saved';
  let lastError = null;

  function drain() {
    if (inFlight) return inFlight;
    inFlight = Promise.resolve().then(async () => {
      while (pending !== null) {
        const value = pending;
        pending = null;
        status = 'saving';
        onStatus(status);
        try {
          await write(value);
          status = pending !== null ? 'dirty' : 'saved';
          lastError = null;
          onStatus(status);
        } catch (error) {
          if (pending === null) pending = value;
          status = 'error';
          lastError = error;
          onStatus(status, error);
          throw error;
        }
      }
    }).finally(() => { inFlight = null; });
    return inFlight;
  }

  return {
    schedule(value) {
      pending = value;
      status = 'dirty';
      onStatus(status);
      queueMicrotask(() => { void drain().catch(() => {}); });
    },
    flush: drain,
    getStatus: () => status,
    getError: () => lastError
  };
}
