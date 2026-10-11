(() => {
  if (window.__qiantieSelectedSubmitPayloadGuardInstalled) return;
  window.__qiantieSelectedSubmitPayloadGuardInstalled = true;

  let lastSelectedIds = [];

  function normalizeIds(value) {
    return [...new Set((Array.isArray(value) ? value : [])
      .map(id => String(id || '').trim())
      .filter(Boolean))];
  }

  function checkedTaskIds() {
    return normalizeIds([...document.querySelectorAll('.task-check:checked')]
      .map(input => input.dataset?.id));
  }

  function rememberCurrentSelection() {
    const ids = checkedTaskIds();
    if (ids.length) lastSelectedIds = ids;
    return ids;
  }

  document.addEventListener('change', event => {
    if (event.target?.matches?.('.task-check')) rememberCurrentSelection();
  }, true);

  document.addEventListener('click', event => {
    if (event.target?.closest?.('#v78QuickSubmit, #openWebSubmitBtn')) rememberCurrentSelection();
  }, true);

  const originalFetch = window.fetch.bind(window);
  window.fetch = async function qiantieSelectedSubmitPayloadGuard(input, init = {}) {
    const url = typeof input === 'string' ? input : String(input?.url || '');
    const isSubmit = /\/api\/batch-rewrite\/web-submit\/submit(?:\?|$)/.test(url);
    if (!isSubmit || typeof init?.body !== 'string') return originalFetch(input, init);

    try {
      const payload = JSON.parse(init.body);
      if (payload?.mode === 'selected' && Array.isArray(payload.ids) && normalizeIds(payload.ids).length === 0) {
        const current = checkedTaskIds();
        const fallback = current.length ? current : lastSelectedIds;
        if (fallback.length) {
          init = { ...init, body: JSON.stringify({ ...payload, ids: fallback }) };
        }
      }
    } catch (_) {
      // Preserve the original request when the body is not JSON.
    }

    return originalFetch(input, init);
  };
})();
