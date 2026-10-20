import { getAgentAssetBlob } from '../../shared/api/agent.js';

export function createAuthenticatedMediaStore({
  loadBlob = getAgentAssetBlob,
  createObjectUrl = blob => URL.createObjectURL(blob),
  revokeObjectUrl = url => URL.revokeObjectURL(url)
} = {}) {
  const entries = new Map();

  function removeEntry(source, entry) {
    if (entries.get(source) === entry) entries.delete(source);
  }

  return {
    acquire(value) {
      const source = String(value || '').trim();
      if (!source) return { promise: Promise.resolve(''), release() {} };

      let entry = entries.get(source);
      if (!entry) {
        entry = { references: 0, objectUrl: '', promise: null };
        entry.promise = Promise.resolve()
          .then(() => loadBlob(source))
          .then(blob => {
            const objectUrl = createObjectUrl(blob);
            entry.objectUrl = objectUrl;
            if (entry.references === 0) {
              revokeObjectUrl(objectUrl);
              entry.objectUrl = '';
              removeEntry(source, entry);
            }
            return objectUrl;
          })
          .catch(error => {
            removeEntry(source, entry);
            throw error;
          });
        entries.set(source, entry);
      }

      entry.references += 1;
      let released = false;
      return {
        promise: entry.promise,
        release() {
          if (released) return;
          released = true;
          entry.references = Math.max(0, entry.references - 1);
          if (entry.references === 0 && entry.objectUrl) {
            revokeObjectUrl(entry.objectUrl);
            entry.objectUrl = '';
            removeEntry(source, entry);
          }
        }
      };
    }
  };
}

export const agentCanvasMediaStore = createAuthenticatedMediaStore();

export function createAuthenticatedMediaSession(source, { store = agentCanvasMediaStore, onState } = {}) {
  if (!source) {
    onState?.({ url: '', loading: false, error: '' });
    return { reportPlaybackError() {}, dispose() {} };
  }

  let active = true;
  const lease = store.acquire(source);
  const release = () => {
    if (!active) return false;
    active = false;
    lease.release();
    return true;
  };
  onState?.({ url: '', loading: true, error: '' });
  lease.promise
    .then(url => {
      if (active) onState?.({ url, loading: false, error: '' });
    })
    .catch(error => {
      if (active) onState?.({ url: '', loading: false, error: error?.message || '视频读取失败' });
    });

  return {
    reportPlaybackError(message = '视频源无法播放') {
      if (!release()) return;
      onState?.({ url: '', loading: false, error: String(message || '').trim() || '视频源无法播放' });
    },
    dispose() {
      release();
    }
  };
}
