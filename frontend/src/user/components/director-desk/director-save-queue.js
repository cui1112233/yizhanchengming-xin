export function createDirectorSaveQueue(write, { onStatus = () => {} } = {}) {
  let pending = null;
  let inFlight = null;
  let status = 'saved';
  let lastError = null;
  let scheduleVersion = 0;

  function drain() {
    if (inFlight) return inFlight;
    let failedAtVersion = null;
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
          failedAtVersion = scheduleVersion;
          onStatus(status, error);
          throw error;
        }
      }
    }).finally(() => {
      inFlight = null;
      if (failedAtVersion !== null && scheduleVersion > failedAtVersion) {
        queueMicrotask(() => { void drain().catch(() => {}); });
      }
    });
    return inFlight;
  }

  return {
    schedule(value) {
      scheduleVersion += 1;
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
